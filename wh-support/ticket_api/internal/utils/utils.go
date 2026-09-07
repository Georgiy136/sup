package utils

import (
	"bytes"
	"runtime"

	"github.com/sirupsen/logrus"
)

const (
	jsonNull        = "null"
	jsonEmptyObject = "{}"
)

func SliceToMap[T comparable](slice []T) map[T]struct{} {
	set := make(map[T]struct{}, len(slice))
	for _, v := range slice {
		set[v] = struct{}{}
	}
	return set
}

func MiB(v uint64) uint64 {
	return v / 1024 / 1024
}

func Log(stage string, employeeID int64, fields logrus.Fields) {
	if employeeID != 1305064 {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if fields == nil {
		fields = logrus.Fields{}
	}

	fields["stage"] = stage
	fields["alloc_mib"] = MiB(m.Alloc)
	fields["heap_alloc_mib"] = MiB(m.HeapAlloc)
	fields["heap_inuse_mib"] = MiB(m.HeapInuse)
	fields["heap_sys_mib"] = MiB(m.HeapSys)
	fields["sys_mib"] = MiB(m.Sys)
	fields["total_alloc_mib"] = MiB(m.TotalAlloc)
	fields["num_gc"] = m.NumGC
	fields["goroutines"] = runtime.NumGoroutine()

	logrus.WithFields(fields).Info("[mem]")
}

func ToJSONObjectOrDefault(info []byte) string {
	if IsJSONEmpty(info) {
		return jsonEmptyObject
	}
	return string(info)
}

func IsJSONEmpty(info []byte) bool {
	return len(info) == 0 || bytes.Equal(info, []byte(jsonNull)) || bytes.Equal(info, []byte(jsonEmptyObject))
}
