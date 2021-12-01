package dtmgimp

import "testing"

func TestGetServerAndMethod(t *testing.T) {
	type args struct {
		grpcURL string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "IpPort",
			args: args{grpcURL: "1.1.1.1:8080/package.service/method"},
			want: "1.1.1.1:8080",
			want1: "/package.service/method",
		},
		{
			name: "Polaris",
			args: args{grpcURL: "polaris://grpc.service/package.service/method?namespace=Production"},
			want: "polaris://grpc.service?namespace=Production",
			want1: "/package.service/method",
		},
		{
			name: "Dns",
			args: args{grpcURL: "dns://host/package.service/method"},
			want: "dns://host",
			want1: "/package.service/method",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetServerAndMethod(tt.args.grpcURL)
			if got != tt.want {
				t.Errorf("GetServerAndMethod() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetServerAndMethod() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
