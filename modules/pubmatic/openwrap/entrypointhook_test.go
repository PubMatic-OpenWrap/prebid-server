package openwrap

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestOpenWrap_handleEntrypointHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	defer ctrl.Finish()

	type fields struct {
		cfg   config.Config
		cache cache.Cache
	}
	type args struct {
		in0     context.Context
		miCtx   hookstage.ModuleInvocationContext
		payload hookstage.EntrypointPayload
		setup   func(*mock_metrics.MockMetricsEngine)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    hookstage.HookResult[hookstage.EntrypointPayload]
		wantErr error
	}{
		{
			name: "request with sshb=1 should not execute entrypointhook",
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5?sshb=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						Sshb: "1",
					},
				},
			},
		},
		{
			name: "valid /openrtb2/2.5 request(reverseProxy through SSHB(8001->8000))",
			fields: fields{
				cfg: config.Config{
					Tracker: config.Tracker{
						Endpoint:                  "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
					},
				},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5?debug=1&sshb=2", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						ProfileID:                 5890,
						DisplayID:                 1,
						DisplayVersionID:          1,
						SSAuction:                 -1,
						Debug:                     true,
						UA:                        "go-test",
						IP:                        "127.0.0.1",
						IsCTVRequest:              false,
						TrackerEndpoint:           "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:             "go-test",
						Aliases:                  make(map[string]string),
						ImpBidCtx:                make(map[string]models.ImpCtx),
						PrebidBidderCode:         make(map[string]string),
						BidderResponseTimeMillis: make(map[string]int),
						ProfileIDStr:             "5890",
						Endpoint:                 models.EndpointV25,
						MetricsEngine:            mockEngine,
						SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid /openrtb/2.5 request with wiid set and no cookies(reverseProxy through SSHB(8001->8000)",
			fields: fields{
				cfg: config.Config{
					Tracker: config.Tracker{
						Endpoint:                  "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
					},
				},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5?debug=1&sshb=2", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1,"wiid":"4df09505-d0b2-4d70-94d9-dc41e8e777f7"}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						ProfileID:                 5890,
						DisplayID:                 1,
						DisplayVersionID:          1,
						SSAuction:                 -1,
						Debug:                     true,
						UA:                        "go-test",
						IP:                        "127.0.0.1",
						IsCTVRequest:              false,
						TrackerEndpoint:           "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
						LoggerImpressionID:        "4df09505-d0b2-4d70-94d9-dc41e8e777f7",
						Aliases:                   make(map[string]string),
						ImpBidCtx:                 make(map[string]models.ImpCtx),
						PrebidBidderCode:          make(map[string]string),
						BidderResponseTimeMillis:  make(map[string]int),
						ProfileIDStr:              "5890",
						Endpoint:                  models.EndpointV25,
						MetricsEngine:             mockEngine,
						SeatNonBids:               make(map[string][]openrtb_ext.NonBid),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "/openrtb/2.5 request without profileid(reverseProxy through SSHB(8001->8000)",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5?&sshb=2", nil)
						if err != nil {
							panic(err)
						}
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileids":5890,"versionid":1}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {
					mme.EXPECT().RecordBadRequests(gomock.Any(), 700)
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidProfileID,
				Errors:  []string{"ErrMissingProfileID"},
			},
			wantErr: nil,
		},
		{
			name: "Valid /openrtb2/auction?source=pbjs request(ows2s)",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=pbjs&debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"imp":[{"tagid":"/43743431/DMDemo","ext":{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890"}},"adunitcode":"div-gpt-ad-1460505748561-0"}},"id":"div-gpt-ad-1460505748561-0","banner":{"topframe":1,"format":[{"w":300,"h":250}]}}],"site":{"domain":"localhost:9999","publisher":{"domain":"localhost:9999","id":"5890"},"page":"http://localhost:9999/integrationExamples/gpt/owServer_example.html"},"device":{"w":1792,"h":446,"dnt":0,"ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36","language":"en","sua":{"source":1,"platform":{"brand":"macOS"},"browsers":[{"brand":"Google Chrome","version":["117"]},{"brand":"Not;A=Brand","version":["8"]},{"brand":"Chromium","version":["117"]}],"mobile":0}},"ext":{"prebid":{"auctiontimestamp":1697191822565,"targeting":{"includewinners":true,"includebidderkeys":true},"bidderparams":{"pubmatic":{"publisherId":"5890","wrapper":{"profileid":43563,"versionid":1}}},"channel":{"name":"pbjs","version":"v8.7.0-pre"},"createtids":false}},"id":"5bdd7da5-1166-40fe-a9cb-3bf3c3164cd3","test":0,"tmax":3000}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						ProfileID:        43563,
						PubID:            0,
						PubIDStr:         "",
						DisplayID:        1,
						DisplayVersionID: 1,
						SSAuction:        -1,
						Debug:            true,
						UA:               "go-test",
						IP:               "127.0.0.1",
						IsCTVRequest:     false,
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:             "go-test",
						Aliases:                  make(map[string]string),
						ImpBidCtx:                make(map[string]models.ImpCtx),
						PrebidBidderCode:         make(map[string]string),
						BidderResponseTimeMillis: make(map[string]int),
						ProfileIDStr:             "43563",
						Endpoint:                 models.EndpointOWS2S,
						MetricsEngine:            mockEngine,
						SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Valid /openrtb2/auction?source=pbjs request(ows2s) debug set from request body instead of queryparam",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=pbjs", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"imp":[{"tagid":"/43743431/DMDemo","ext":{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890"}},"adunitcode":"div-gpt-ad-1460505748561-0"}},"id":"div-gpt-ad-1460505748561-0","banner":{"topframe":1,"format":[{"w":300,"h":250}]}}],"site":{"domain":"localhost:9999","publisher":{"domain":"localhost:9999","id":"5890"},"page":"http://localhost:9999/integrationExamples/gpt/owServer_example.html"},"device":{"w":1792,"h":446,"dnt":0,"ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36","language":"en","sua":{"source":1,"platform":{"brand":"macOS"},"browsers":[{"brand":"Google Chrome","version":["117"]},{"brand":"Not;A=Brand","version":["8"]},{"brand":"Chromium","version":["117"]}],"mobile":0}},"ext":{"prebid":{"auctiontimestamp":1697191822565,"targeting":{"includewinners":true,"includebidderkeys":true},"bidderparams":{"pubmatic":{"publisherId":"5890","wrapper":{"profileid":43563,"versionid":1}}},"channel":{"name":"pbjs","version":"v8.7.0-pre"},"createtids":false,"debug":true}},"id":"5bdd7da5-1166-40fe-a9cb-3bf3c3164cd3","test":0,"tmax":3000}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						ProfileID:        43563,
						PubID:            0,
						PubIDStr:         "",
						DisplayID:        1,
						DisplayVersionID: 1,
						SSAuction:        -1,
						Debug:            true,
						UA:               "go-test",
						IP:               "127.0.0.1",
						IsCTVRequest:     false,
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:             "go-test",
						Aliases:                  make(map[string]string),
						ImpBidCtx:                make(map[string]models.ImpCtx),
						PrebidBidderCode:         make(map[string]string),
						BidderResponseTimeMillis: make(map[string]int),
						ProfileIDStr:             "43563",
						Endpoint:                 models.EndpointOWS2S,
						MetricsEngine:            mockEngine,
						SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "/openrtb2/auction?source=pbjs request(ows2s) without profileid",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=pbjs&debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"imp":[{"tagid":"/43743431/DMDemo","ext":{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890"}},"adunitcode":"div-gpt-ad-1460505748561-0"}},"id":"div-gpt-ad-1460505748561-0","banner":{"topframe":1,"format":[{"w":300,"h":250}]}}],"site":{"domain":"localhost:9999","publisher":{"domain":"localhost:9999","id":"5890"},"page":"http://localhost:9999/integrationExamples/gpt/owServer_example.html"},"device":{"w":1792,"h":446,"dnt":0,"ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36","language":"en","sua":{"source":1,"platform":{"brand":"macOS"},"browsers":[{"brand":"Google Chrome","version":["117"]},{"brand":"Not;A=Brand","version":["8"]},{"brand":"Chromium","version":["117"]}],"mobile":0}},"ext":{"prebid":{"auctiontimestamp":1697191822565,"targeting":{"includewinners":true,"includebidderkeys":true},"bidderparams":{"pubmatic":{"publisherId":"5890","wrapper":{}}},"channel":{"name":"pbjs","version":"v8.7.0-pre"},"createtids":false}},"id":"5bdd7da5-1166-40fe-a9cb-3bf3c3164cd3","test":0,"tmax":3000}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {
					mme.EXPECT().RecordBadRequests(gomock.Any(), 700)
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidProfileID,
				Errors:  []string{"ErrMissingProfileID"},
			},
			wantErr: nil,
		},
		{
			name: "/openrtb2/auction without source=pbjs/inapp. new hybrid endpoint should not execute module",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"imp":[{"tagid":"/43743431/DMDemo","ext":{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890"}},"adunitcode":"div-gpt-ad-1460505748561-0"}},"id":"div-gpt-ad-1460505748561-0","banner":{"topframe":1,"format":[{"w":300,"h":250}]}}],"site":{"domain":"localhost:9999","publisher":{"domain":"localhost:9999","id":"5890"},"page":"http://localhost:9999/integrationExamples/gpt/owServer_example.html"},"device":{"w":1792,"h":446,"dnt":0,"ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36","language":"en","sua":{"source":1,"platform":{"brand":"macOS"},"browsers":[{"brand":"Google Chrome","version":["117"]},{"brand":"Not;A=Brand","version":["8"]},{"brand":"Chromium","version":["117"]}],"mobile":0}},"ext":{"prebid":{"auctiontimestamp":1697191822565,"targeting":{"includewinners":true,"includebidderkeys":true},"bidderparams":{"pubmatic":{"publisherId":"5890","wrapper":{}}},"channel":{"name":"pbjs","version":"v8.7.0-pre"},"createtids":false}},"id":"5bdd7da5-1166-40fe-a9cb-3bf3c3164cd3","test":0,"tmax":3000}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						Endpoint: models.EndpointHybrid,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "hybrid api pbs/openrtb2/auction should not execute entrypointhook",
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/pbs/openrtb2/auction", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						Endpoint: models.EndpointHybrid,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid /openrtb2/auction?source=inapp request directly on port 8000",
			fields: fields{
				cfg: config.Config{
					Tracker: config.Tracker{
						Endpoint:                  "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
					},
				},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=inapp&debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						ProfileID:                 5890,
						DisplayID:                 1,
						DisplayVersionID:          1,
						SSAuction:                 -1,
						Debug:                     true,
						UA:                        "go-test",
						IP:                        "127.0.0.1",
						IsCTVRequest:              false,
						TrackerEndpoint:           "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:             "go-test",
						Aliases:                  make(map[string]string),
						ImpBidCtx:                make(map[string]models.ImpCtx),
						PrebidBidderCode:         make(map[string]string),
						BidderResponseTimeMillis: make(map[string]int),
						ProfileIDStr:             "5890",
						Endpoint:                 models.EndpointV25,
						MetricsEngine:            mockEngine,
						SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.setup(mockEngine)
			m := OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: mockEngine,
			}
			got, err := m.handleEntrypointHook(tt.args.in0, tt.args.miCtx, tt.args.payload)
			assert.Equal(t, err, tt.wantErr)

			if tt.want.ModuleContext != nil {
				// validate runtime values individually and reset them
				gotRctx := got.ModuleContext["rctx"].(models.RequestCtx)
				if gotRctx.Endpoint == models.EndpointHybrid || gotRctx.Sshb == "1" {
					assert.Equal(t, got, tt.want)
					return
				}

				assert.NotEmpty(t, gotRctx.StartTime)
				gotRctx.StartTime = 0

				wantRctx := tt.want.ModuleContext["rctx"].(models.RequestCtx)
				if wantRctx.LoggerImpressionID == "" {
					assert.Len(t, gotRctx.LoggerImpressionID, 36)
					gotRctx.LoggerImpressionID = ""
				}
				gotRctx.ParsedUidCookie = nil    // ignore parsed cookies
				gotRctx.CurrencyConversion = nil // ignore currency-conversion
				got.ModuleContext["rctx"] = gotRctx
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
