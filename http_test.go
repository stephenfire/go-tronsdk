package go_tronsdk

import (
	"context"
	"testing"
	"time"
)

const contentType = "application/json"

// func TestWebConn(t *testing.T) {
// 	client := new(http.Client)
// 	data := "{\"address\": \"415134e32fb878d2d4eeb1bdfce1fda42389575722\"}"
// 	req, err := http.NewRequestWithContext(
// 		context.Background(),
// 		http.MethodPost,
// 		"https://nile.trongrid.io/wallet/getaccount",
// 		io.NopCloser(bytes.NewReader([]byte(data))))
// 	req.Header = make(http.Header)
// 	req.Header.Set("accept", contentType)
// 	req.Header.Set("content-type", contentType)
// 	req.ContentLength = int64(len(data))
//
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer func() {
// 		_ = resp.Body.Close()
// 	}()
// 	var acc core.Account
// 	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
// 		t.Fatal(fmt.Errorf("failed request: %d", resp.StatusCode))
// 	}
// 	buf := new(bytes.Buffer)
// 	if _, err = buf.ReadFrom(resp.Body); err != nil {
// 		t.Fatal(err)
// 	}
// 	body := buf.Bytes()
// 	if err = json.Unmarshal(body, &acc); err != nil {
// 		t.Errorf("%s", string(body))
// 		t.Fatal(err)
// 	}
// 	t.Log(acc)
// }

func TestHttpClient_GetNextMaintenanceTime(t *testing.T) {
	client := NewHttpClient("https://nile.trongrid.io", 15)
	defer client.Close()
	ti, err := client.GetNextMaintenanceTime(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("next time: %s, %s", ti.String(), ti.In(time.UTC))
}
