package localization

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func TestLocalization(t *testing.T) {
	testcases := []struct {
		testName     string
		messages     []LocalizationKey
		expectedJSON string
	}{
		{
			testName: "Сообщение с одним ключом-значением",
			messages: []LocalizationKey{
				{KeyID: "std.std.err_biz.employee.not_found",
					MessageValues: map[string]string{"employee_id": "12345"}},
			},
			expectedJSON: `{
			  "keys_with_values": [
			    {
			      "key_id": "std.std.err_biz.employee.not_found",
			      "message_values": {
			        "employee_id": "12345"
			      }
			    }
			  ]
			}`,
		},
		{
			testName: "Сообщения с разными ключами",
			messages: []LocalizationKey{
				{
					KeyID:         "std.std.err_biz.place_type.wrong.required",
					MessageValues: map[string]string{"place_type_id": "233"},
				},
				{
					KeyID:         "std.std.err_biz.employee.not_found",
					MessageValues: map[string]string{"employee_id": "12345"},
				},
				{
					KeyID:         "auth.std.err_biz.token.wh.wrong",
					MessageValues: map[string]string{"current_wh_id": "212", "expected_wh_id": "34"},
				},
				{
					KeyID:         "invent.std.err_biz.task.already_exists",
					MessageValues: map[string]string{"task_id": "999", "wh_id": "55"},
				},
				{
					KeyID:         "support.std.err_biz.street.invalid_range",
					MessageValues: map[string]string{"start": "10", "end": "5"},
				},
			},
			expectedJSON: `{
			  "keys_with_values": [
			    {
			      "key_id": "std.std.err_biz.place_type.wrong.required",
			      "message_values": {
			        "place_type_id": "233"
			      }
			    },
			    {
			      "key_id": "std.std.err_biz.employee.not_found",
			      "message_values": {
			        "employee_id": "12345"
			      }
			    },
			    {
			      "key_id": "auth.std.err_biz.token.wh.wrong",
			      "message_values": {
			        "current_wh_id": "212",
			        "expected_wh_id": "34"
			      }
			    },
			    {
			      "key_id": "invent.std.err_biz.task.already_exists",
			      "message_values": {
			        "task_id": "999",
			        "wh_id": "55"
			      }
			    },
			    {
			      "key_id": "support.std.err_biz.street.invalid_range",
			      "message_values": {
			        "start": "10",
			        "end": "5"
			      }
			    }
			  ]
			}`,
		},
		{
			testName: "Сообщения с множеством параметров",
			messages: []LocalizationKey{
				{
					KeyID: "support.std.err_biz.invent_task.create_failed",
					MessageValues: map[string]string{
						"stage":         "5",
						"section_start": "100",
						"section_end":   "200",
						"wh_id":         "42",
					},
				},
				{
					KeyID: "support.std.err_biz.storage_place.create_failed",
					MessageValues: map[string]string{
						"place_id":   "777",
						"place_type": "shelf",
						"zone":       "A1",
					},
				},
				{
					KeyID: "support.std.err_biz.tare.exclude_failed",
					MessageValues: map[string]string{
						"tare_id":    "12345",
						"tare_type":  "box",
						"reason":     "occupied",
						"wh_id":      "99",
						"updated_at": "2024-01-15",
					},
				},
				{
					KeyID: "support.std.err_biz.wh.close_failed",
					MessageValues: map[string]string{
						"wh_id":  "101",
						"status": "active",
					},
				},
				{
					KeyID: "support.std.err_biz.stage.create_failed",
					MessageValues: map[string]string{
						"stage_id":   "15",
						"stage_name": "picking",
						"wh_id":      "42",
						"error_code": "ERR_001",
					},
				},
			},
			expectedJSON: `{
			  "keys_with_values": [
			    {
			      "key_id": "support.std.err_biz.invent_task.create_failed",
			      "message_values": {
			        "stage": "5",
			        "section_start": "100",
			        "section_end": "200",
			        "wh_id": "42"
			      }
			    },
			    {
			      "key_id": "support.std.err_biz.storage_place.create_failed",
			      "message_values": {
			        "place_id": "777",
			        "place_type": "shelf",
			        "zone": "A1"
			      }
			    },
			    {
			      "key_id": "support.std.err_biz.tare.exclude_failed",
			      "message_values": {
			        "tare_id": "12345",
			        "tare_type": "box",
			        "reason": "occupied",
			        "wh_id": "99",
			        "updated_at": "2024-01-15"
			      }
			    },
			    {
			      "key_id": "support.std.err_biz.wh.close_failed",
			      "message_values": {
			        "wh_id": "101",
			        "status": "active"
			      }
			    },
			    {
			      "key_id": "support.std.err_biz.stage.create_failed",
			      "message_values": {
			        "stage_id": "15",
			        "stage_name": "picking",
			        "wh_id": "42",
			        "error_code": "ERR_001"
			      }
			    }
			  ]
			}`,
		},
		{
			testName: "Смешанные сообщения: nil, пустой map и с данными",
			messages: []LocalizationKey{
				{
					KeyID:         "support.std.err_biz.ticket.not_found",
					MessageValues: nil,
				},
				{
					KeyID:         "support.std.wrn_biz.operation.skipped",
					MessageValues: map[string]string{},
				},
				{
					KeyID:         "support.std.err_biz.shk.limit_exceeded",
					MessageValues: map[string]string{"limit": "1000", "actual": "1500"},
				},
				{
					KeyID:         "support.std.inf_biz.task.completed",
					MessageValues: nil,
				},
				{
					KeyID:         "support.std.wrn_biz.validation.empty_field",
					MessageValues: map[string]string{"field_name": "description"},
				},
			},
			expectedJSON: `{
			  "keys_with_values": [
			    {
			      "key_id": "support.std.err_biz.ticket.not_found",
			      "message_values": null
			    },
			    {
			      "key_id": "support.std.wrn_biz.operation.skipped",
			      "message_values": {}
			    },
			    {
			      "key_id": "support.std.err_biz.shk.limit_exceeded",
			      "message_values": {
			        "limit": "1000",
			        "actual": "1500"
			      }
			    },
			    {
			      "key_id": "support.std.inf_biz.task.completed",
			      "message_values": null
			    },
			    {
			      "key_id": "support.std.wrn_biz.validation.empty_field",
			      "message_values": {
			        "field_name": "description"
			      }
			    }
			  ]
			}`,
		},
		{
			testName: "Сообщение с пустым значением",
			messages: []LocalizationKey{
				{KeyID: "std.std.wrn_biz.request.limit", MessageValues: map[string]string{}},
			},
			expectedJSON: `{
			  "keys_with_values": [
			    {
			      "key_id": "std.std.wrn_biz.request.limit",
			      "message_values": {}
			    }
			  ]
			}`,
		},
		{
			testName: "Сообщение с nil значением",
			messages: []LocalizationKey{
				{KeyID: "invent.std.err_biz.access.denied", MessageValues: nil},
			},
			expectedJSON: `{
			  "keys_with_values": [
			    {
			      "key_id": "invent.std.err_biz.access.denied",
			      "message_values": null
			    }
			  ]
			}`,
		},
		{
			testName: "Пустой LocalizedErrors",
			messages: []LocalizationKey{},
			expectedJSON: `{
			  "keys_with_values": null
			}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			actual := &LocalizedErrors{}
			for _, msg := range tc.messages {
				actual.Add(msg.KeyID, msg.MessageValues)
			}

			var expected LocalizedErrors
			if err := jsoniter.UnmarshalFromString(tc.expectedJSON, &expected); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}

			if len(expected.KeysWithValues) != len(actual.KeysWithValues) {
				t.Fatalf("Length mismatch: expected %d, got %d", len(expected.KeysWithValues), len(actual.KeysWithValues))
			}

			for i := range expected.KeysWithValues {
				exp := expected.KeysWithValues[i]
				act := actual.KeysWithValues[i]

				// Сравниваем KeyID
				if exp.KeyID != act.KeyID {
					t.Errorf("[%d]: KeyID mismatch: expected %q, got %q", i, exp.KeyID, act.KeyID)
				}

				// Сравниваем MessageValues
				if !mapsEqual(exp.MessageValues, act.MessageValues) {
					t.Errorf("[%d]: MessageValues mismatch: expected %v, got %v", i, exp.MessageValues, act.MessageValues)
				}
			}
		})
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}
