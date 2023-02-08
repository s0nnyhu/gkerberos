package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/itchyny/volume-go"
	"github.com/rs/cors"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Output  string `json:output`
}

type Cmd struct {
	Cmd string `json:"cmd"`
}

//go:embed static/index.html
var indexPage []byte

//go:embed static/assets
var assets embed.FS

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {
	ip := GetOutboundIP()
	r := mux.NewRouter()

	// Handle API routes
	api := r.PathPrefix("/api/v1/").Subrouter()
	api.HandleFunc("/commands", handleCommands).Methods("POST", "OPTIONS")

	// Serve static files
	fsys := fs.FS(assets)
	assetsStatic, _ := fs.Sub(fsys, "static/assets")
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.FS(assetsStatic))))
	// Serve index page on all unhandled routes
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(indexPage)
	})

	log.Printf("Listening on %s:8080", ip.String())

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	// start server listen
	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe(ip.String()+":8080", handler))

}

func handleCommands(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "NA"
	message := "NA"

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	var data Cmd
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Println(w, "err: %v", err)
		status = "KO"
		message = "err:"

		response := &Response{
			Status:  status,
			Message: message,
			Output:  "",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
	}

	if strings.Contains(data.Cmd, "volume") {
		handleVolumeCommand(data, w)
	} else {
		handleCmdCommand(data, w)
	}

}

func handleVolumeCommand(data Cmd, w http.ResponseWriter) {
	status := "NA"
	message := "NA"

	_, err := volume.GetVolume()

	s := strings.Split(data.Cmd, " ")

	value, errValue := strconv.Atoi(s[1])
	if errValue != nil {
		if err != nil {
			status = "KO"
			message = "set volume failed " + err.Error()
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("set volume failed: %+v", err)
		}
	} else {
		err = volume.SetVolume(value)
		status = "OK"
		message = "set volume success"
		w.WriteHeader(http.StatusCreated)
		log.Printf("set volume success\n")
	}

	response := &Response{
		Status:  status,
		Message: message,
		Output:  "",
	}

	json.NewEncoder(w).Encode(response)
}

func handleCmdCommand(data Cmd, w http.ResponseWriter) {
	status := "NA"
	message := "NA"

	out, err := exec.Command("cmd", "/C", data.Cmd).Output()

	if err != nil {
		log.Println("Failed to initiate cmd:", err)
		status = "KO"
		message = "Failed to initiate cmd: " + err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Println("Executing " + data.Cmd)
		status = "OK"
		message = "cmd executed successfully"
		w.WriteHeader(http.StatusCreated)
	}

	// if err := exec.Command("cmd", "/C", data.Cmd).Run(); err != nil {
	// 	log.Println("Failed to initiate cmd:", err)
	// 	status = "KO"
	// 	message = "Failed to initiate cmd: " + err.Error()
	// 	w.WriteHeader(http.StatusInternalServerError)
	// } else {
	// 	log.Println("Executing " + data.Cmd)
	// 	status = "OK"
	// 	message = "cmd executed successfully"
	// 	w.WriteHeader(http.StatusCreated)
	// }

	response := &Response{
		Status:  status,
		Message: message,
		Output:  string(out),
	}

	json.NewEncoder(w).Encode(response)
}
