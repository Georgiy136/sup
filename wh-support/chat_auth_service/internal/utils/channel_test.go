package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTicketIDFromChannel(t *testing.T) {
	type args struct {
		channel string
	}
	tests := []struct {
		name      string
		args      args
		want      int64
		wantError bool
	}{
		{
			name: "Канал заявки",
			args: args{
				channel: "ticket:12345",
			},
			want:      12345,
			wantError: false,
		},
		{
			name: "Канал чата заявки",
			args: args{
				channel: "ticket:12345:chat",
			},
			want:      12345,
			wantError: false,
		},
		{
			name: "Неверный формат канала",
			args: args{
				channel: "invalid-channel",
			},
			want:      0,
			wantError: true,
		},
		{
			name: "Отсутствует ticketID",
			args: args{
				channel: "ticket:",
			},
			want:      0,
			wantError: true,
		},
		{
			name: "Нечисловой ticket ID",
			args: args{
				channel: "ticket:abc",
			},
			want:      0,
			wantError: true,
		},
		{
			name: "Нулевой ticket ID",
			args: args{
				channel: "ticket:0",
			},
			want:      0,
			wantError: true,
		},
		{
			name: "Отрицательный ticketID",
			args: args{
				channel: "ticket:-123",
			},
			want:      0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTicketIDFromChannel(tt.args.channel)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
