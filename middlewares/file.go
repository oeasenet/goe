package middlewares

import (
	"crypto/md5"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/strutil"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.oease.dev/goe/core"
	"go.oease.dev/goe/models"
	"go.oease.dev/goe/storages/s3minio"
	"go.oease.dev/goe/utils"
	"go.oease.dev/goe/webresult"
	"io"
	"strings"
	"sync"
	"time"
)

type FileMiddlewares struct {
	storage *s3minio.Storage
	cfg     *FileMiddlewareConfig
}

var defaultAllowedMimeTypes = []string{
	"application/epub+zip",
	"application/gzip",
	"application/java-archive",
	"application/json",
	"application/ld+json",
	"application/msword",
	"application/octet-stream",
	"application/ogg",
	"application/pdf",
	"application/rtf",
	"application/vnd.amazon.ebook",
	"application/vnd.apple.installer+xml",
	"application/vnd.mozilla.xul+xml",
	"application/vnd.ms-excel",
	"application/vnd.ms-fontobject",
	"application/vnd.ms-powerpoint",
	"application/vnd.oasis.opendocument.presentation",
	"application/vnd.oasis.opendocument.spreadsheet",
	"application/vnd.oasis.opendocument.text",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/vnd.rar",
	"application/vnd.visio",
	"application/x-7z-compressed",
	"application/x-abiword",
	"application/x-bzip",
	"application/x-bzip2",
	"application/x-csh",
	"application/x-freearc",
	"application/x-sh",
	"application/x-shockwave-flash",
	"application/x-tar",
	"application/xhtml+xml",
	"application/xml",
	"application/zip",
	"audio/aac",
	"audio/midi",
	"audio/mpeg",
	"audio/ogg",
	"audio/opus",
	"audio/wav",
	"audio/webm",
	"audio/3gpp",
	"audio/3gpp2",
	"font/otf",
	"font/ttf",
	"font/woff",
	"font/woff2",
	"image/bmp",
	"image/gif",
	"image/jpeg",
	"image/png",
	"image/svg+xml",
	"image/tiff",
	"image/webp",
	"text/calendar",
	"text/css",
	"text/csv",
	"text/html",
	"text/javascript",
	"text/plain",
	"text/xml",
	"video/3gpp",
	"video/3gpp2",
	"video/mp2t",
	"video/mp4",
	"video/mpeg",
	"video/ogg",
	"video/webm",
	"video/x-msvideo",
}

type FileMiddlewareConfig struct {
	// MaxUploadLimit is the maximum file size allowed for upload in bytes. Default is 200MB.
	UploadLimit int64

	// UploadFormKey is the key used to access the uploaded file in the form data. Default is "files".
	UploadFormKey string

	// AllowedMimeTypes is a list of allowed MIME types for file upload. Default all common file types.
	AllowedMimeTypes []string

	// FileIdRouteKey is the key used to access the file ID in the route parameters. Default is "id".
	IdRouteKey string

	// HashRouteKey is the hash used to access the file hash in the route parameters. Default is "hash".
	HashRouteKey string
}

var DefaultFileMiddlewareConfig = FileMiddlewareConfig{
	UploadLimit:      209715200,
	UploadFormKey:    "files",
	AllowedMimeTypes: defaultAllowedMimeTypes,
	IdRouteKey:       "id",
	HashRouteKey:     "hash",
}

func NewFileMiddlewares(config ...FileMiddlewareConfig) *FileMiddlewares {
	bucket := core.UseGoeConfig().S3.Bucket
	if bucket == "" {
		panic("S3 bucket is not configured")
		return nil
	}
	endpoint := core.UseGoeConfig().S3.Endpoint
	if endpoint == "" {
		panic("S3 endpoint is not configured")
		return nil
	}
	region := core.UseGoeConfig().S3.Region
	if region == "" {
		panic("S3 region is not configured")
		return nil
	}
	bucketLookup := s3minio.BucketLookupAuto
	if core.UseGoeConfig().S3.BucketLookup == "dns" {
		bucketLookup = s3minio.BucketLookupDNS
	} else if core.UseGoeConfig().S3.BucketLookup == "path" {
		bucketLookup = s3minio.BucketLookupPath
	} else {
		bucketLookup = s3minio.BucketLookupAuto
	}
	accessKey := core.UseGoeConfig().S3.AccessKey
	secretKey := core.UseGoeConfig().S3.SecretKey
	if accessKey == "" || secretKey == "" {
		panic("S3 access key or secret key is not configured")
		return nil
	}
	secure := core.UseGoeConfig().S3.UseSSL
	store := s3minio.New(s3minio.Config{
		Bucket:       bucket,
		Endpoint:     endpoint,
		Region:       region,
		BucketLookup: bucketLookup,
		Token:        core.UseGoeConfig().S3.Token,
		Secure:       secure,
		Reset:        false,
		Credentials: s3minio.Credentials{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretKey,
		},
		GetObjectOptions:    minio.GetObjectOptions{},
		PutObjectOptions:    minio.PutObjectOptions{},
		ListObjectsOptions:  minio.ListObjectsOptions{},
		RemoveObjectOptions: minio.RemoveObjectOptions{},
	})
	if len(config) == 0 {
		return &FileMiddlewares{
			storage: store,
			cfg:     &DefaultFileMiddlewareConfig,
		}
	}
	if config[0].UploadLimit == 0 {
		config[0].UploadLimit = DefaultFileMiddlewareConfig.UploadLimit
	}
	if config[0].UploadFormKey == "" {
		config[0].UploadFormKey = DefaultFileMiddlewareConfig.UploadFormKey
	}
	if len(config[0].AllowedMimeTypes) == 0 {
		config[0].AllowedMimeTypes = DefaultFileMiddlewareConfig.AllowedMimeTypes
	}
	return &FileMiddlewares{
		storage: store,
		cfg:     &config[0],
	}
}

// HandleUpload handles the file upload request.
// Route recommendation: POST /file/upload
// It parses the multipart form and retrieves all files from the "files" key.
// Then, it performs form validation on each file, checking the size and MIME type.
// The function processes each file, calculating its hash, determining the ideal filename,
// and saving the file to the storage and its info to the database.
// The function returns an error if the multipart form data is invalid or any other error occurs.
func (m *FileMiddlewares) HandleUpload() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		// Parse the multipart form:
		if form, err := ctx.MultipartForm(); err == nil {
			// Get all files from "files" key:
			files := form.File["files"]

			if len(files) == 0 {
				return webresult.SendFailed(ctx, "no file to upload")
			}

			// Form validation, loop through files
			for _, file := range files {
				//check size, then check mime type
				if file.Size > m.cfg.UploadLimit {
					return webresult.SendFailed(ctx, fmt.Sprintf("file size too large (max %s)", utils.ConvertBytesToHumanReadableSize(int(m.cfg.UploadLimit))))
				}
				// Check if the file MIME type is allowed
				mimeType := file.Header.Get("Content-Type")
				if !utils.ArrContainsStr(m.cfg.AllowedMimeTypes, mimeType) {
					return webresult.SendFailed(ctx, fmt.Sprintf("%s file type is not allowed", mimeType))
				}
			}

			fileInfos := make([]*models.GoeFile, 0)
			// Loop through files and process upload:
			for _, file := range files {
				// Calculate and check all the file properties
				// Open the file, so we can read the content and calculate the hash
				openedFile, err := file.Open()
				if err != nil {
					return webresult.SystemBusy(err)
				}
				// Calculate the file hash
				hasher := md5.New()
				if _, err := copyIOZeroAlloc(hasher, openedFile); err != nil {
					return webresult.SystemBusy(err)
				}
				fileHash := fmt.Sprintf("%x", hasher.Sum(nil))
				_ = openedFile.Close()

				// check if file with same hash exist, if exist return the file info no more upload needed
				fileInfo := &models.GoeFile{}
				hasResult, err := core.UseGoeContainer().GetMongo().FindOne(fileInfo, bson.M{"hash": fileHash}, fileInfo)
				if hasResult {
					fileInfos = append(fileInfos, fileInfo)
					continue
				}

				fileExtWDot := fsutil.FileExt(file.Filename)
				fileExtWODot := strings.TrimPrefix(fileExtWDot, ".")

				//calculate ideal uploaded filename
				idealFileName := strutil.Md5(fmt.Sprintf(`%d|%s|%s`, time.Now().UnixMilli(), fileHash, file.Filename)) + fileExtWDot

				dbFileType := determineFileTypeFromExt(fileExtWDot)

				fileInfo = &models.GoeFile{
					Extension:    fileExtWODot,
					Filename:     file.Filename,
					Hash:         fileHash,
					MimeType:     file.Header.Get("Content-Type"),
					Size:         file.Size,
					Type:         dbFileType,
					UploadedName: idealFileName,
				}

				// Save the file to storage
				if err := ctx.SaveFileToStorage(file, fmt.Sprintf("./%s", idealFileName), m.storage); err != nil {
					return webresult.SystemBusy(err)
				}

				// Save the file info to database
				_, err = core.UseGoeContainer().GetMongo().Insert(fileInfo)
				if err != nil {
					return webresult.SystemBusy(err)
				}
				fileInfos = append(fileInfos, fileInfo)
			}
			return webresult.SendSucceed(ctx, fileInfos)
		} else {
			return webresult.InvalidParam("invalid multipart form data")
		}
	}
}

// HandleView handles the file view request.
// Route recommendation: GET /file/view/:id
// It checks if the request is for downloading the file and sets the appropriate download flag.
// Then, it gets the file ID from the route and finds the file info from the database.
// It retrieves the file content from the storage and sends it as the response.
// If the download flag is set, it sets the Content-Disposition header to suggest downloading.
// The function returns an error if the file is not found or any other error occurs.
func (m *FileMiddlewares) HandleView() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		//if request download
		mustDownload := false
		if ctx.Query("download", "false") == "true" {
			mustDownload = true
		} else {
			mustDownload = false
		}

		//if request download with custom name
		downloadName := ""
		if mustDownload {
			dNameQuery := ctx.Query("name")
			if dNameQuery != "" {
				downloadName = dNameQuery
			}
		}

		//get file id from route
		fileId := ctx.Params(m.cfg.IdRouteKey)
		if fileId == "" {
			return webresult.InvalidParam("invalid file id")
		}

		//find file info from database
		fileInfo := &models.GoeFile{}
		hasResult, err := core.UseGoeContainer().GetMongo().FindById(fileInfo, fileId, fileInfo)
		if !hasResult {
			return webresult.NotFound("file not found")
		}
		if err != nil {
			return webresult.SystemBusy(err)
		}

		//get file content from storage and display it
		fileData, err := m.storage.Get(fmt.Sprintf("./%s", fileInfo.UploadedName))
		if err != nil {
			return webresult.SystemBusy(err)
		}
		if fileData == nil {
			return webresult.NotFound("file not found in upstream storage")
		}
		ctx.Response().Header.SetContentType(fileInfo.MimeType)
		ctx.Response().SetStatusCode(fiber.StatusOK)

		if mustDownload {
			if downloadName != "" {
				ctx.Response().Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadName))
			} else {
				ctx.Response().Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Filename))
			}
		}

		return ctx.Send(fileData)
	}
}

// HandleDelete handles the file delete request.
// Route recommendation: DELETE /file/delete/:id
// It gets the file ID from the route and validates it.
// Then, it finds the file info from the database using the ID.
// Next, it deletes the file from the storage by providing the file path.
// Finally, it deletes the file info from the database.
// The function returns an error if the file ID is invalid, file is not found,
// there's an error deleting the file from storage, or error deleting the file info from the database.
func (m *FileMiddlewares) HandleDelete() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		//get file id from route
		fileId := ctx.Params(m.cfg.IdRouteKey)
		if fileId == "" {
			return webresult.InvalidParam("invalid file id")
		}

		//find file info from database
		fileInfo := &models.GoeFile{}
		hasResult, err := core.UseGoeContainer().GetMongo().FindById(fileInfo, fileId, fileInfo)
		if !hasResult {
			return webresult.NotFound("file not found")
		}
		if err != nil {
			return webresult.SystemBusy(err)
		}

		//delete file from storage
		err = m.storage.Delete(fmt.Sprintf("./%s", fileInfo.UploadedName))
		if err != nil {
			return webresult.SystemBusy(err)
		}

		//delete file info from database
		err = core.UseGoeContainer().GetMongo().Delete(fileInfo)
		if err != nil {
			return webresult.SystemBusy(err)
		}

		return webresult.SendSucceed(ctx)
	}
}

// HandleMatch handles the file match request.
// Route recommendation: GET /file/match/:hash
// It gets the file MD5 hash from the route parameter.
// Then, it finds the file info from the database using the hash.
// The function returns the file info if found, otherwise it returns an error.
func (m *FileMiddlewares) HandleMatch() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		//get file id from route
		hash := ctx.Params(m.cfg.HashRouteKey)
		if hash == "" {
			return webresult.InvalidParam("invalid file md5 hash")
		}

		//find file info from database
		fileInfo := &models.GoeFile{}
		hasResult, err := core.UseGoeContainer().GetMongo().FindOne(fileInfo, bson.M{"hash": hash}, fileInfo)
		if !hasResult {
			return webresult.NotFound("hash not found")
		}
		if err != nil {
			return webresult.SystemBusy(err)
		}

		return webresult.SendSucceed(ctx, fileInfo)
	}
}

// determineFileTypeFromExt determines the file type based on the provided file extension.
// It maps the file extension to the corresponding FileType enum value.
// If the file extension is not recognized, it defaults to FileTypeOther.
// The function is case-insensitive when comparing the file extension.
// Example usage:
// fileType := determineFileTypeFromExt(".jpg")
func determineFileTypeFromExt(fileExt string) models.FileType {
	fileType := models.FileTypeOther
	switch strings.ToLower(fileExt) {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tif", ".tiff", ".svg":
		fileType = models.FileTypeImage
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm", ".m4v", ".3gp", ".mpg", ".mpeg":
		fileType = models.FileTypeVideo
	case ".mp3", ".wav", ".wma", ".ogg", ".flac", ".aac", ".m4a", ".alac", ".aiff":
		fileType = models.FileTypeAudio
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".pdf", ".txt", ".odt", ".ods", ".odp", ".rtf", ".md":
		fileType = models.FileTypeDocument
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".iso":
		fileType = models.FileTypeArchive
	case ".exe", ".msi", ".apk", ".ipa", ".dmg", ".pkg", ".deb", ".rpm", ".bat", ".sh", ".jar", ".py", ".js":
		fileType = models.FileTypeApplication
	default:
		fileType = models.FileTypeOther
	}
	return fileType
}

// copyBufPool is a sync.Pool that provides a pool of byte slices for copying data.
// It is used in the `CopyIOZeroAlloc` function to efficiently allocate and reuse byte slices.
// The New field of copyBufPool is set to a function that creates a new byte slice of size 4096.
// Example usage:
//
//	vbuf := copyBufPool.Get()
//	buf := vbuf.([]byte)
//	n, err := io.CopyBuffer(w, r, buf)
//	copyBufPool.Put(vbuf)
//	return n, err
var copyBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

// copyIOZeroAlloc copies data from a reader to a writer using a zero-allocation buffer pool.
// It retrieves a byte slice from the copyBufPool and uses it as a buffer for efficient copying.
// The buffer is reused by returning it to the pool after the copy operation is complete.
// The function returns the number of bytes copied and any error encountered during copying.
// Example usage:
//
//	n, err := CopyIOZeroAlloc(w, r)
func copyIOZeroAlloc(w io.Writer, r io.Reader) (int64, error) {
	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)
	n, err := io.CopyBuffer(w, r, buf)
	copyBufPool.Put(vbuf)
	return n, err
}
