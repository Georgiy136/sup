package bandaction

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
)

var defaultURLRegexp = regexp.MustCompile(consts.DefaultURLPattern)

func (b *BandActionService) CanAttachBandActions(data *ticketmodels.CategoryStatusData) bool {
	if data == nil {
		logrus.Error("[canProcessInteractiveDialog] categoryStatusData is nil")
		return false
	}
	if hasFieldsToFill(data) {
		_, err := convertFieldsToDialogElements(data.AddedInformationModel)
		return err == nil
	}
	return true
}

func hasFieldsToFill(data *ticketmodels.CategoryStatusData) bool {
	if data == nil {
		logrus.Error("[hasFieldsToFill] categoryStatusData is nil")
		return false
	}
	for _, scenarios := range data.AddedInformationModel {
		if len(scenarios.ScenarioAddedInfoModel) > 0 {
			return true
		}
	}
	return false
}

func convertFieldsToDialogElements(scenaries []ticketmodels.ScenarioInfo) ([]models.BandDialogElement, error) {
	if len(scenaries) != 1 {
		return nil, fmt.Errorf("multiply scenaries not supported, scenaries len=%d ", len(scenaries))
	}
	scenarioAddedInfoModel := scenaries[0].ScenarioAddedInfoModel

	sort.Slice(scenarioAddedInfoModel, func(i, j int) bool {
		return scenarioAddedInfoModel[i].OrderID < scenarioAddedInfoModel[j].OrderID
	})

	elements := make([]models.BandDialogElement, 0, len(scenarioAddedInfoModel))

	for _, field := range scenarioAddedInfoModel {
		el, err := convertFieldToDialogElement(&field)
		if err != nil {
			return nil, fmt.Errorf("error converting scenario field for '%s': %w", field.FrontDataName, err)
		}
		elements = append(elements, *el)
	}

	return elements, nil
}

func convertFieldToDialogElement(field *ticketmodels.ScenarioField) (*models.BandDialogElement, error) {
	if field == nil {
		return nil, errors.New("scenario field is nil")
	}
	el := models.BandDialogElement{
		DisplayName: field.FrontDataName,
		Name:        field.DataName,
		Required:    field.IsRequired,
	}

	switch field.FrontDataType {
	case ticketmodels.FieldTypeSelector:
		el.Type = models.BandDialogElementTypeSelect
		if field.IsMultiple {
			return nil, errors.New("multiple selectors not supported")
		}
		if field.Api != nil {
			return nil, errors.New("api for selectors not supported")
		}
		if len(field.SelectorArray) == 0 {
			return nil, errors.New("selector array len is 0")
		}

		selectorArray := make([]ticketmodels.SelectorOption, len(field.SelectorArray))
		copy(selectorArray, field.SelectorArray)
		sort.Slice(selectorArray, func(i, j int) bool {
			return selectorArray[i].OrderID < selectorArray[j].OrderID
		})

		for _, opt := range selectorArray {
			var valueID string
			switch v := opt.ID.(type) {
			case float64:
				valueID = strconv.FormatFloat(v, 'f', -1, 64)
			case string:
				valueID = v
			default:
				return nil, fmt.Errorf("can't parse id in selector array: %+v. id not a string or float64, got %+v", selectorArray, opt.ID)
			}
			el.Options = append(el.Options, &models.BandDialogElementOption{
				Text:  opt.Name,
				Value: valueID,
			})
		}
	case ticketmodels.FieldTypeDate:
		el.Type = models.BandDialogElementTypeText
		el.Placeholder = consts.DatePlaceholder

	case ticketmodels.FieldTypeBoolean:
		el.Type = models.BandDialogElementTypeBool
		el.Required = false
		el.Placeholder = field.FrontDataDescription

	case ticketmodels.FieldTypeNumber:
		el.Type = models.BandDialogElementTypeText
		el.Placeholder = field.FrontDataDescription

	case ticketmodels.FieldTypeString:
		el.Type = models.BandDialogElementTypeText
		el.Placeholder = field.FrontDataDescription

		if field.LongText {
			el.Type = models.BandDialogElementTypeTextarea
		}
		if len(field.DefaultValues.Values) > 0 {
			switch v := field.DefaultValues.Values[0].(type) {
			case float64:
				el.DefaultValue = strconv.FormatFloat(v, 'f', -1, 64)
			case string:
				el.DefaultValue = v
			default:
				return nil, fmt.Errorf("default_value not a string or float64, got %+v", field.DefaultValues.Values)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported type '%s'", field.FrontDataType)
	}

	return &el, nil
}

func (b *BandActionService) validateAndConvertSubmissionValues(ctx context.Context, categoryID int64, statusID string, submission map[string]any) (map[string]any, error) {
	categoryStatusResp, err := b.ticketRepo.GetCategoryStatusByStatus(ctx, categoryID, statusID)
	if err != nil {
		return nil, fmt.Errorf("can't get category status for value conversion: %w", err)
	}
	if len(categoryStatusResp.Data) != 1 {
		return nil, errors.New("get category status error, data len > 1")
	}
	if len(categoryStatusResp.Data[0].AddedInformationModel) != 1 {
		return nil, errors.New("get category status error, scenarios len > 1")
	}

	result := make(map[string]any, len(submission))
	validationErrors := make(map[string]string, len(categoryStatusResp.Data[0].AddedInformationModel))

	for _, field := range categoryStatusResp.Data[0].AddedInformationModel[0].ScenarioAddedInfoModel {
		rawVal, exists := submission[field.DataName]
		if !exists {
			if !field.IsRequired {
				continue
			}
			return nil, fmt.Errorf("field '%s' doesn't exist in submission", field.DataName)
		}
		converted, err := convertFieldValue(field, rawVal)
		if err != nil {
			validationErr, ok := errors.AsType[*customerrors.SubmissionFieldValidationError](err)
			if !ok {
				return nil, fmt.Errorf("can't convert field '%s', error: %w", field.DataName, err)
			}
			validationErrors[field.DataName] = validationErr.Error()
		}

		result[field.DataName] = converted
	}
	if len(validationErrors) > 0 {
		return nil, customerrors.NewInteractiveDialogValidationError(validationErrors)
	}

	return result, nil
}

func convertFieldValue(field ticketmodels.ScenarioField, value any) (any, error) {
	switch field.FrontDataType {
	case ticketmodels.FieldTypeNumber:
		if value == nil {
			return nil, fmt.Errorf("can't convert field '%s' with nil value", field.FrontDataName)
		}
		strVal, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("can't convert field '%s', not string value", field.FrontDataName)
		}
		if !field.IsRequired && strVal == "" {
			return nil, nil
		}

		numVal, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			if numErr, ok := errors.AsType[*strconv.NumError](err); ok && errors.Is(numErr.Err, strconv.ErrRange) {
				return nil, customerrors.NewSubmissionFieldValidationError(consts.ValidationErrorIntegerTooLarge)
			}
			return nil, customerrors.NewSubmissionFieldValidationError(consts.ValidationErrorIntegerOnly)
		}
		if field.MinThreshold != nil {
			val, err := strconv.ParseInt(*field.MinThreshold, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error converting min threshold for '%s': %w", field.FrontDataName, err)
			}
			if numVal < val {
				return nil, customerrors.NewSubmissionFieldValidationError(fmt.Sprintf(consts.ValidationErrorMinValueFormat, val))
			}
		}
		if field.MaxThreshold != nil {
			val, err := strconv.ParseInt(*field.MaxThreshold, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error converting max threshold for '%s': %w", field.FrontDataName, err)
			}
			if numVal > val {
				return nil, customerrors.NewSubmissionFieldValidationError(fmt.Sprintf(consts.ValidationErrorMaxValueFormat, val))
			}
		}
		return numVal, nil
	case ticketmodels.FieldTypeBoolean:
		if value == nil {
			return nil, fmt.Errorf("can't convert field '%s' with nil value", field.FrontDataName)
		}
		boolVal, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("can't convert field '%s', not bool value", field.FrontDataName)
		}
		return boolVal, nil
	case ticketmodels.FieldTypeSelector:
		if value == nil {
			return nil, fmt.Errorf("can't convert field '%s' with nil value", field.FrontDataName)
		}
		strVal, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("can't convert field '%s', not string value", field.FrontDataName)
		}
		if !field.IsRequired && strVal == "" {
			return nil, nil
		}

		var (
			valueName string
			valueID   any
		)
		for _, opt := range field.SelectorArray {
			var valueIDStr string
			switch v := opt.ID.(type) {
			case float64:
				valueIDStr = strconv.FormatFloat(v, 'f', -1, 64)
			case string:
				valueIDStr = v
			default:
				return nil, fmt.Errorf("can't parse id in selector array: %+v. id not a string or float64, got %+v", field.SelectorArray, opt.ID)
			}
			if valueIDStr == strVal {
				valueName = opt.Name
				valueID = opt.ID
				break
			}
		}
		if valueName == "" {
			return nil, fmt.Errorf("can't find valueName for valueID '%s'", strVal)
		}

		val := map[string]any{
			"id":   valueID,
			"name": valueName,
		}
		return val, nil
	case ticketmodels.FieldTypeDate:
		if value == nil {
			return nil, fmt.Errorf("can't convert field '%s' with nil value", field.FrontDataName)
		}
		strVal, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("can't convert field '%s', not string value", field.FrontDataName)
		}
		if !field.IsRequired && strVal == "" {
			return nil, nil
		}

		const layout = "02.01.2006"

		parsed, err := time.Parse(layout, strVal)
		if err != nil {
			return nil, customerrors.NewSubmissionFieldValidationError(consts.DatePlaceholder)
		}
		return parsed.Format(layout), nil
	case ticketmodels.FieldTypeString:
		if value == nil {
			return nil, fmt.Errorf("can't convert field '%s' with nil value", field.FrontDataName)
		}
		strVal, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("can't convert field '%s', not string value", field.FrontDataName)
		}
		if !field.IsRequired && strVal == "" {
			return nil, nil
		}

		if err := validateFieldPattern(field, strVal); err != nil {
			return nil, fmt.Errorf("validate input by pattern for '%s': %w", field.FrontDataName, err)
		}

		strLen := utf8.RuneCountInString(strVal)
		if field.MinThreshold != nil {
			val, err := strconv.ParseInt(*field.MinThreshold, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error converting min threshold for '%s': %w", field.FrontDataName, err)
			}
			if strLen < int(val) {
				return nil, customerrors.NewSubmissionFieldValidationError(fmt.Sprintf(consts.ValidationErrorMinLengthFormat, val))
			}
		}
		if field.MaxThreshold != nil {
			val, err := strconv.ParseInt(*field.MaxThreshold, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error converting max threshold for '%s': %w", field.FrontDataName, err)
			}
			if strLen > int(val) {
				return nil, customerrors.NewSubmissionFieldValidationError(fmt.Sprintf(consts.ValidationErrorMaxLengthFormat, val))
			}
		}
		return strVal, nil
	}
	return nil, fmt.Errorf("unsupported type '%s'", field.FrontDataType)
}

func validateFieldPattern(field ticketmodels.ScenarioField, value string) error {
	switch {
	case field.RegularExpression != nil && field.RegularExpression.RegExp != "":
		re, err := regexp.Compile(field.RegularExpression.RegExp)
		if err != nil {
			return fmt.Errorf("compile regular expression: %w", err)
		}
		if !re.MatchString(value) {
			return customerrors.NewSubmissionFieldValidationError(field.RegularExpression.WarnMessage)
		}
	case field.IsURL:
		if !defaultURLRegexp.MatchString(value) {
			return customerrors.NewSubmissionFieldValidationError(consts.ValidationErrorLinkRequired)
		}
	}
	return nil
}
