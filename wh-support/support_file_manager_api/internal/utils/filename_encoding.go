package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// generateShards вычисляет shardId и shardId2 из хэша идентификатора.
// shardId - первые 2 символа хэша, shardId2 - следующие 2 символа.
func generateShards(id int64) (shardId, shardId2 string) {
	hash := md5.Sum([]byte(strconv.Itoa(int(id))))
	hashStr := hex.EncodeToString(hash[:])
	return hashStr[:2], hashStr[2:4]
}

// GenerateTmpObjectKey генерирует ключ для временного хранилища.
// Формат: tmp/{shardId}/{shardId2}/{uploadId}
// Пример: tmp/a1/b2/12345
func GenerateTmpObjectKey(uploadID int64) string {
	shardId, shardId2 := generateShards(uploadID)
	return fmt.Sprintf("tmp/%s/%s/%d", shardId, shardId2, uploadID)
}

// GenerateFinalObjectKey генерирует ключ для постоянного хранилища.
// Формат: final/{entityType}/{shard1}/{shard2}/{fileId}
func GenerateFinalObjectKey(entityType string, fileId int64) string {
	shardOne, shardTwo := generateShards(fileId)
	return fmt.Sprintf("final/%s/%s/%s/%d", entityType, shardOne, shardTwo, fileId)
}

func EncodeFilenameForMetadata(filename string) string {
	return url.PathEscape(filename)
}

func DecodeFilenameFromMetadata(encodedFilename string) string {
	decoded, err := url.PathUnescape(encodedFilename)
	if err != nil {
		return encodedFilename
	}
	return decoded
}

func GetContentDisposition(contentType string) string {
	if strings.HasPrefix(contentType, "image/") {
		return "inline"
	}
	return "attachment"
}
