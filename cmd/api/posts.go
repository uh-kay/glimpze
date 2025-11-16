package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uh-kay/glimpze/store"
)

const maxFileSize = 4 << 20
const maxFormSize = 4<<20 + 8192

var ErrUnsupportedFile = errors.New("file must be jpg, jpeg, png, gif, or webp")

type Form struct {
	Content string `json:"content" validate:"required,min=1,max=2048"`
}

func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := r.ParseMultipartForm(maxFormSize); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	content := r.PostFormValue("content")
	if err := Validate.Struct(Form{Content: content}); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) > 4 {
		app.badRequestResponse(w, r, errors.New("you can only upload 4 files per post"))
		return
	}

	for _, fileHeader := range files {
		if err := app.validateFileUpload(fileHeader); err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
	}

	var post *store.Post
	var err error
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		post, err = s.Posts.Create(r.Context(), content, user.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	postFileRecords := make([]any, 0, len(files))

	for _, fileHeader := range files {
		postFileRecord, _, err := app.processFileUpload(r.Context(), fileHeader, post.ID)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		postFileRecords = append(postFileRecords, postFileRecord)
	}

	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		err = app.store.UserLimits.Reduce(r.Context(), user.ID, "create_post_limit")
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		Message: "post created",
		Data: map[string]any{
			"post":  post,
			"files": postFileRecords,
		},
	})
}

type PostFileWithLink struct {
	PostFile   *store.PostFile `json:"post_file"`
	PublicLink string          `json:"public_link"`
}

func (app *application) getPost(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post, err := app.store.Posts.GetByID(r.Context(), postID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	var postFiles []*store.PostFile
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		postFiles, err = app.store.PostFiles.GetByPostID(r.Context(), post.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	fileLinks := make(map[string]string, len(postFiles))

	for _, val := range postFiles {
		publicLink, err := app.storage.GetFromR2(r.Context(), fmt.Sprintf("%s%s", val.FileID, val.FileExtension))
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		fileLinks[val.FileID.String()] = publicLink
	}

	postFileWithLinks := make([]PostFileWithLink, 0, len(postFiles))
	for i := range len(postFiles) {
		postFileWithLink := PostFileWithLink{
			PostFile:   postFiles[i],
			PublicLink: fileLinks[postFiles[i].FileID.String()],
		}
		postFileWithLinks = append(postFileWithLinks, postFileWithLink)
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		Message: "success",
		Data: map[string]any{
			"post":      post,
			"post_file": postFileWithLinks,
		},
	})
}

func (app *application) updatePost(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("postID")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := r.ParseMultipartForm(maxFormSize); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	content := r.PostFormValue("content")
	if err := Validate.Struct(Form{Content: content}); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	files := r.MultipartForm.File["file"]
	if len(files) > 4 {
		app.badRequestResponse(w, r, errors.New("you can only upload 4 files per post"))
		return
	}
	for _, fileHeader := range files {
		if err := app.validateFileUpload(fileHeader); err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
	}

	var post *store.Post
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		post, err = app.store.Posts.Update(r.Context(), content, int64(postID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var postFileRecords []any
	var oldPostFiles []*store.PostFile

	if len(files) > 0 {
		err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
			oldPostFiles, err = app.store.PostFiles.GetByPostID(r.Context(), post.ID)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		postFileRecords = make([]any, 0, len(files))
		filenames := make([]string, 0, len(files))

		for _, fileHeader := range files {
			postFileRecord, filename, err := app.processFileUpload(r.Context(), fileHeader, post.ID)
			if err != nil {
				app.cleanupUploadedFiles(r.Context(), filenames)
				app.internalServerError(w, r, err)
				return
			}
			postFileRecords = append(postFileRecords, postFileRecord)
			filenames = append(filenames, filename)
		}

		for _, val := range oldPostFiles {
			err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
				err = app.store.PostFiles.Delete(r.Context(), val.FileID)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				app.cleanupUploadedFiles(r.Context(), filenames)
				app.internalServerError(w, r, err)
				return
			}
		}

		for _, val := range oldPostFiles {
			err := app.storage.DeleteFromR2(r.Context(), fmt.Sprintf("%s%s", val.FileID, val.FileExtension))
			if err != nil {
				app.logger.Error("failed to delete old file", "error", err)
			}
		}
	} else {
		err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
			oldPostFiles, err = app.store.PostFiles.GetByPostID(r.Context(), post.ID)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		postFileRecords = make([]any, len(oldPostFiles))
		for i, f := range oldPostFiles {
			postFileRecords[i] = f
		}
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		Message: "post updated",
		Data: map[string]any{
			"post":      post,
			"post_file": postFileRecords,
		},
	})
}

func (app *application) deletePost(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("postID")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var postFiles []*store.PostFile
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		postFiles, err = app.store.PostFiles.GetByPostID(r.Context(), int64(postID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	for _, val := range postFiles {
		err = app.store.WithTx(r.Context(), func(s *store.Storage) error {

			err = app.store.PostFiles.Delete(r.Context(), val.FileID)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

	}

	if err = app.store.Posts.Delete(r.Context(), int64(postID)); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	for _, val := range postFiles {
		err := app.storage.DeleteFromR2(r.Context(), fmt.Sprintf("%s%s", val.FileID, val.FileExtension))
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) validateFileUpload(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > maxFileSize {
		return errors.New("file size exceeds 4MB limit")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		return ErrUnsupportedFile
	}

	contentType, err := app.getContentType(fileHeader)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(contentType, "image/") {
		return ErrUnsupportedFile
	}

	return nil
}

func (app *application) getContentType(fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	return http.DetectContentType(buffer[:n]), nil
}

func (app *application) processFileUpload(ctx context.Context, fileHeader *multipart.FileHeader, postID int64) (*store.PostFile, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	fileID := uuid.New()
	fileExt := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s%s", fileID.String(), fileExt)

	if err := app.storage.SaveToR2(ctx, file, fileExt, filename); err != nil {
		return nil, "", err
	}

	var postFile *store.PostFile
	err = app.store.WithTx(ctx, func(s *store.Storage) error {
		postFile, err = app.store.PostFiles.Create(ctx, fileID, fileExt, fileHeader.Filename, postID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	return postFile, filename, nil
}

func (app *application) cleanupUploadedFiles(ctx context.Context, filenames []string) {
	for _, val := range filenames {
		_ = app.storage.DeleteFromR2(ctx, val)
	}
}

func getPostFromContext(r *http.Request) *store.Post {
	return r.Context().Value(postCtx).(*store.Post)
}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postIDStr := r.PathValue("postID")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		post, err := app.store.Posts.GetByID(r.Context(), postID)
		if err != nil {
			switch err {
			case pgx.ErrNoRows:
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) addLike(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	post := getPostFromContext(r)

	var err error
	var postLike *store.PostLike
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		postLike, err = app.store.PostLikes.Create(r.Context(), user.ID, post.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "post_likes_pkey":
				app.badRequestResponse(w, r, ErrDuplicateLike)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		Message: "like added",
		Data:    postLike,
	})
}

func (app *application) removeLike(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	post := getPostFromContext(r)

	var err error
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		err = app.store.PostLikes.Delete(r.Context(), user.ID, post.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
