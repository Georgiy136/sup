package docgen

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

type (
	gotenbergClient interface {
		ConvertRenderTemplateToPdf(ctx context.Context, docxData []byte) ([]byte, error)
	}
)

type Service struct {
	gotenbergClient gotenbergClient
}

func NewService(gotenbergClient gotenbergClient) *Service {
	return &Service{
		gotenbergClient: gotenbergClient,
	}
}

// RenderTemplateToPdf принимает шаблон DOCX и подстановки, возвращает PDF.
func (s *Service) RenderTemplateToPdf(ctx context.Context, docxBytes []byte, replacements map[string]string) ([]byte, error) {
	var err error
	if len(replacements) > 0 {
		docxBytes, err = renderDocx(docxBytes, replacements)
		if err != nil {
			return nil, fmt.Errorf("can't render docx: %w", err)
		}
	}

	pdfBytes, err := s.gotenbergClient.ConvertRenderTemplateToPdf(ctx, docxBytes)
	if err != nil {
		return nil, fmt.Errorf("can't convert docx to pdf: %w", err)
	}

	return pdfBytes, nil
}

// renderDocx подставляет значения в DOCX-шаблон и возвращает готовый DOCX.
func renderDocx(templateBytes []byte, replacements map[string]string) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(templateBytes), int64(len(templateBytes)))
	if err != nil {
		return nil, fmt.Errorf("can't open docx as zip: %w", err)
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, file := range zipReader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("can't open %s: %w", file.Name, err)
		}

		content, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("can't read %s: %w", file.Name, err)
		}

		if err = rc.Close(); err != nil {
			logrus.Errorf("can't close docx: %v", err)
		}

		if isWordXML(file.Name) {
			content = replacePlaceholders(content, replacements)
		}

		w, err := zipWriter.Create(file.FileHeader.Name)
		if err != nil {
			return nil, fmt.Errorf("can't create %s: %w", file.Name, err)
		}
		if _, err = w.Write(content); err != nil {
			return nil, fmt.Errorf("can't write %s: %w", file.Name, err)
		}
	}

	if err = zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("can't close docx: %w", err)
	}

	return buf.Bytes(), nil
}

func isWordXML(name string) bool {
	return name == "word/document.xml" ||
		strings.HasPrefix(name, "word/footer") ||
		strings.HasPrefix(name, "word/header")
}

func replacePlaceholders(content []byte, replacements map[string]string) []byte {
	for key, value := range replacements {
		content = bytes.ReplaceAll(content, []byte(key), []byte(value))
	}
	return content
}
