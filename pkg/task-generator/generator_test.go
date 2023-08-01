package generator

import (
	"reflect"
	"testing"
	"time"
)

func TestDecryptCronExpress(t *testing.T) {
	type args struct{ cronExpr string }
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{name: "Test case 1",
			args: args{cronExpr: "0 0 * * * *"},
			want: time.Date(time.Now().Year(), time.Now().Month(),
				time.Now().Day(), time.Now().Hour()+1, 0, 0, 0, time.Local)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecryptCronExpress(tt.args.cronExpr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecryptCronExpress() = %v, want %v", got, tt.want)
			}
		})
	}
}
