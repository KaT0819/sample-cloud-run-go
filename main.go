package main

import (
	"encoding/json"
	"fmt"
	"hello-run/snssms"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gorilla/sessions"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// templateData provides template parameters.
type templateData struct {
	Service  string
	Revision string
}

// Variables used to generate the HTML page.
var (
	data templateData
	tmpl *template.Template
)

type Temps struct {
	notemp *template.Template
	index  *template.Template
	hello  *template.Template
	login  *template.Template
}

var cs *sessions.CookieStore = sessions.NewCookieStore([]byte("secret-key-12345"))

// Template for no-template
func notemp() *template.Template {
	src := "<html></html>"
	tmp, _ := template.New("index").Parse(src)
	return tmp
}

// setup template function
func setupTemp() *Temps {
	temps := new(Temps)

	temps.notemp = notemp()

	// set index template.
	index, err := template.ParseFiles("templates/signin.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		index = temps.notemp
	}
	temps.index = index

	// set hello template
	hello, err := template.ParseFiles("templates/hello.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		hello = temps.notemp
	}
	temps.hello = hello

	// set hello template
	login, err := template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		login = temps.notemp
	}
	temps.login = login

	return temps
}

// index handler
func index(w http.ResponseWriter, r *http.Request, tmp *template.Template) {
	content := struct {
		Title   string
		Message string
	}{
		Title:   "Index",
		Message: "Index Page",
	}

	err := tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

// hello handler
func hello(w http.ResponseWriter, r *http.Request, tmp *template.Template) {

	// パラメータ取得
	name := r.FormValue("name")

	msg := "template message<br>これはサンプルです。" + name

	if r.Method == "POST" {
		pass := r.FormValue("pass")
		msg = fmt.Sprintf("Name is %s Password is %s", name, pass)
	}

	content := struct {
		Title       string
		Message     string
		Name        string
		Flg         bool
		SubMessage1 string
		SubMessage2 string
		Items       []string
	}{
		Title:       "template title",
		Message:     msg,
		Name:        name,
		Flg:         false,
		SubMessage1: "サブタイトル",
		SubMessage2: "てすと",
		Items:       []string{"あいうえお", "かきくけこ", "１２３４５６７８９０", "!\"#$%&'()", "<b><i><u>htmlタグ</u></i></b>"},
	}
	err := tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

// login handler
func login(w http.ResponseWriter, r *http.Request, tmp *template.Template) {

	ses, err := cs.Get(r, "hello-session")
	if err != nil {
		log.Fatal(err)
	}

	msg := "名前とパスワードを入力してください。"

	if r.Method == "POST" {
		ses.Values["login"] = false
		ses.Values["name"] = nil
		// パラメータ取得
		name := r.FormValue("name")
		pass := r.FormValue("pass")

		if name == pass {
			ses.Values["login"] = true
			ses.Values["name"] = name
		}
		ses.Save(r, w)

		msg = fmt.Sprintf("Name is %s Password is %s", name, pass)
	} else {
		ses.Values["login"] = false
		ses.Values["name"] = nil

		ses.Save(r, w)
	}

	flg, _ := ses.Values["login"].(bool)
	name, _ := ses.Values["name"].(string)

	if flg {
		msg = "logined " + name
	}

	content := struct {
		Title   string
		Message string
	}{
		Title:   "Cookie Session",
		Message: msg,
	}

	err = tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Initialize template parameters.
	service := os.Getenv("K_SERVICE")
	if service == "" {
		service = "???"
	}

	revision := os.Getenv("K_REVISION")
	if revision == "" {
		revision = "???"
	}

	// Prepare template for execution.
	tmpl = template.Must(template.ParseFiles("index.html"))
	data = templateData{
		Service:  service,
		Revision: revision,
	}

	// Define HTTP server.
	http.HandleFunc("/", helloRunHandler)

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	temps := setupTemp()

	// index handle
	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		index(w, r, temps.index)
	})
	// hello handle
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		hello(w, r, temps.hello)
	})
	// login handle
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		login(w, r, temps.login)
	})

	// PORT environment variable is provided by Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Cloud Runにアクセス
	// url := "https://toki-test-run-km6ljd432a-an.a.run.app/"

	// resp, _ := http.Get(url)
	// defer resp.Body.Close()

	// byteArray, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(byteArray)) // htmlをstringで取得

	// AWS SNSにアクセス
	// fsendsms()
	// sendmessage()
	// sendvonage()
	// sendgridmail()

	log.Print("Hello from Cloud Run! The container started successfully and is listening for HTTP requests on $PORT")
	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// helloRunHandler responds to requests by rendering an HTML page.
func helloRunHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		log.Printf("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

func fsendsms() {
	log.Printf("sendsms start")
	// クライアントの生成
	client, err := snssms.GetClient("AKIATD424O6C4MUK7V44", "RqOc9IHdILQ8OrGhxW21/WgdzXYt5+uRxapePeff", "ap-northeast-1")
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}
	log.Printf("sendsms client")

	// SMS送信
	msgIn := snssms.CreateInputMessage("TestMessage", "+818074910888")
	log.Printf("Listening on port %s", msgIn)

	result, err := client.Publish(msgIn)
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}

	fmt.Printf("Result: %s", result.String())
}

const (
	TopicARN = "arn:aws:sns:ap-northeast-1:214536320901:email-topic"
	// TopicARN  = "arn:aws:sns:ap-northeast-1:008023950074:test-toki"
	AwsRegion = "ap-northeast-1"
)

const (
	AppId    = "80fb90a6c28247689bf341c2e141b1e6" // PinpointのプロジェクトID
	SenderId = "rescuenow"                        // 送信ID SMSの送信者名
)

// SNS
func sendSns() {
	log.Printf("sendSns start")
	mySession := session.Must(session.NewSession())
	svc := sns.New(mySession, aws.NewConfig().WithRegion(AwsRegion))

	text := "Amazon SNS サンプルメッセージです。地震が発生しました。"

	// サブスクリプションのプロトコルごとにメッセージを指定
	messageJson := map[string]string{
		"default": text,
		"sqs":     "This is sample message for sqs." + text,
		"sms":     "SMS エンドポイントのサンプルメッセージ" + text,
		"email":   "E メールエンドポイントのサンプルメッセージ" + text,
	}
	// メッセージ構造体はJSON文字列にする
	bytes, err := json.Marshal(messageJson)
	if err != nil {
		fmt.Println("JSON marshal Error: ", err)
	}
	message := string(bytes)

	// inputSms := &sns.PublishInput{
	// 	Message:     aws.String(message),
	// 	PhoneNumber: aws.String("+818074910888"),
	// }

	pin := &sns.PublishInput{}
	pin.SetMessage("SMS エンドポイントのサンプルメッセージ" + text)
	pin.SetPhoneNumber("+818074910888")
	outSms, err := svc.Publish(pin)
	if err != nil {
		fmt.Println("Publish Error: ", err)
	}
	log.Printf(outSms.GoString())

	inputPublish := &sns.PublishInput{
		Message:          aws.String(message),
		MessageStructure: aws.String("json"), // MessageStructureにjsonを指定
		TopicArn:         aws.String(TopicARN),
	}

	MessageId, err := svc.Publish(inputPublish)
	if err != nil {
		fmt.Println("Publish Error: ", err)
	}

	// fmt.Println(MessageId)
	log.Printf(MessageId.GoString())

}

// Pinpoint
func sendmessage() {
	log.Printf("sendmessage start")
	mySession := session.Must(session.NewSession())

	text := "Pinpoint サンプルメッセージです。地震が発生しました。"

	// サブスクリプションのプロトコルごとにメッセージを指定
	messageJson := map[string]string{
		"default": text,
		"sqs":     "This is sample message for sqs." + text,
		"sms":     "SMS エンドポイントのサンプルメッセージ" + text,
		"email":   "E メールエンドポイントのサンプルメッセージ" + text,
	}
	// メッセージ構造体はJSON文字列にする
	bytes, err := json.Marshal(messageJson)
	if err != nil {
		fmt.Println("JSON marshal Error: ", err)
	}
	message := string(bytes)

	pp := pinpoint.New(mySession, aws.NewConfig().WithRegion(AwsRegion))

	// 送信先電話番号
	// phoneNum := "+818074910888"

	// SMS送信
	pIn := &pinpoint.SendMessagesInput{
		ApplicationId: aws.String(AppId),
		MessageRequest: &pinpoint.MessageRequest{
			Addresses: map[string]*pinpoint.AddressConfiguration{
				"+818074910888": &pinpoint.AddressConfiguration{
					ChannelType: aws.String(pinpoint.ChannelTypeSms),
				},
			},
			MessageConfiguration: &pinpoint.DirectMessageConfiguration{
				SMSMessage: &pinpoint.SMSMessage{
					Body:        aws.String(message),  // 本文
					SenderId:    aws.String(SenderId), // 送信ID SMSの送信者名
					MessageType: aws.String(pinpoint.MessageTypePromotional),
				},
			},
		},
	}

	pOut, _ := pp.SendMessages(pIn)
	log.Println(pOut.MessageResponse.Result)
}

// Vonage
func sendvonage() {
	number := "818074910888"
	var API_KEY = "6b4727dd"
	var API_SECRET = "0qtbugeJoLkSxKve"

	value := url.Values{}
	value.Set("from", "rescuenow")
	value.Add("text", "サンプルメッセージ By Vonage API")
	value.Add("to", number)
	value.Add("api_key", API_KEY)
	value.Add("api_secret", API_SECRET)
	value.Add("type", "unicode")
	resp, err := http.PostForm("https://rest.nexmo.com/sms/json", value)
	if err != nil {
		log.Fatal(err)
	}
	buffer := make([]byte, 1024)
	respLen, _ := resp.Body.Read(buffer)
	body := string(buffer[:respLen])
	fmt.Println(body)
	fmt.Println(resp.Status)
	defer resp.Body.Close()
}

// Sendgrid
func sendgridmail() {
	from := mail.NewEmail("Example User", "test@example.com")
	subject := "レスキューナウよりお知らせ Sendgrid"
	to := mail.NewEmail("Example User", "katuyuki.toki@gmail.com")
	plainTextContent := "サンプルテキストメッセージの送信"
	htmlContent := "<strong>サンプルテキストメッセージの送信</strong>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
