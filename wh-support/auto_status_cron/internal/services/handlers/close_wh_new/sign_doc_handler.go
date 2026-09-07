package closewh

import (
	"context"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/consts"

	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

const (
	signDocTemplatePath = "final/FILE_GEN_TEMPLATE/close_wh/v1/writeoff_goods_memo.docx"

	excelMimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	pdfMimeType   = "application/pdf"

	signDocFileNameFormat = "close_block_%d.pdf"
)

func (t *TicketHandlerCloseWhNew) generateSignDoc(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext closewhmodelsnew.ExtTicketInfoForGenerateSignDoc
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("can't decode ext to template data: %w", err)
	}

	if len(t.templateDocx) == 0 {
		return fmt.Errorf("template docx is empty")
	}

	docReplacements, err := buildSignDocReplacements(ticketInfo.TicketID, ext)
	if err != nil {
		return fmt.Errorf("can't get document replacements: %w", err)
	}

	pdfBytes, err := t.docGen.RenderTemplateToPdf(ctx, t.templateDocx, docReplacements)
	if err != nil {
		return fmt.Errorf("can't generate pdf: %w", err)
	}

	documentName := fmt.Sprintf(signDocFileNameFormat, ext.WhId.Id)
	fileID, err := t.fileManagerApi.UploadFile(ctx, pdfBytes, documentName)
	if err != nil {
		return fmt.Errorf("can't upload pdf: %w", err)
	}

	filesJSON, err := jsoniter.Marshal(map[string][]models.Files{
		"generated_file": {{FileID: fileID, FileType: pdfMimeType, FileSize: int64(len(pdfBytes)), FileName: documentName}},
	})
	if err != nil {
		return fmt.Errorf("can't marshal files: %w", err)
	}

	if err = t.ticketApi.PerformTicketV2(ctx, models.RequestPerformTicketV2{
		TicketID:        ticketInfo.TicketID,
		ScenarioOrderID: 0,
		Ext:             filesJSON,
		Files:           filesJSON,
	}, consts.SystemEmployeeID); err != nil {
		return fmt.Errorf("can't perform ticket with files: %w", err)
	}
	return nil
}
