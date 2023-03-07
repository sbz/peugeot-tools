package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	neturl "net/url"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

const (
	ClientID           = "20a4cf7c-f5fb-41d5-9175-a6e23b9880e5" // client_id from Peugeot Update.app/Contents/Resources/app.asar
	VehiculeIdentifier = "VF3CCHNZTHT014827"                    // not mine
	UpdateApiURL       = "https://api.groupe-psa.com/applications/majesticf/v1/getAvailableUpdate"
)

var aliasTypes = map[string]string{
	"ovip-int-firmware-version": "NAC Firmware",
	"map-eur":                   "GPS Map",
	"rcc-firmware":              "RCC Firmware",
}

type UpdateInfoResult struct {
	UpdateId      string `json:"updateId"`
	UpdateSize    string `json:"updateSize"`
	UpdateVersion string `json:"updateVersion"`
	UpdateDate    string `json:"updateDate"`
	UpdateURL     string `json:"updateURL"`
	LicenseURL    string `json:"licenseURL"`
}
type SoftwareResult struct {
	SoftwareType           string             `json:"softwareType"`
	UpdateRequestResult    string             `json:"updateRequestResult"`
	CurrentSoftwareVersion string             `json:"currentSoftwareVersion"`
	Update                 []UpdateInfoResult `json:"update"`
}

type UpdateResult struct {
	RequestResult      string           `json:"requestResult"`
	InstallerURL       string           `json:"installerURL"`
	VehiculeIdentifier string           `json:"vin"`
	Softwares          []SoftwareResult `json:"software"`
}

type SoftwareRequest struct {
	SoftwareType string `json:"softwareType"`
}

type AvailableRequest struct {
	VehiculeIdentifier string            `json:"vin"`
	SoftwareTypes      []SoftwareRequest `json:"softwareTypes"`
}

var jsonPayload string = `{
	"vin": "%s",
	"softwareTypes": [{
			"softwareType": "ovip-int-firmware-version"
		},
		{
			"softwareType": "rcc-firmware"
		},
		{
			"softwareType": "aio-firmware"
		},
		{
			"softwareType": "map-eur"
		}
	]
}`

func createPayload(VehiculeIdentifier string) []byte {
	var payload []byte

	request := AvailableRequest{
		VehiculeIdentifier: VehiculeIdentifier,
		SoftwareTypes: []SoftwareRequest{
			{SoftwareType: "ovip-int-firmware-version"},
			{SoftwareType: "rcc-firmware"},
			{SoftwareType: "map-eur"},
			{SoftwareType: "aio-firmware"},
		},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("oups")
	}

	fmt.Printf("%+v\n", string(payload))

	return payload
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "    ")
	if err != nil {
		return in
	}
	return out.String()
}

func downloadFile(urlpath string) (bool, error) {
	resumeBool := false
	fileOutput := "output.bin"

	fileInfo, err := os.Stat(fileOutput)
	if !os.IsNotExist(err) {
		resumeBool = true
	}

	fd, err := os.OpenFile(fileOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error open " + err.Error())
	}
	defer fd.Close()

	request, nil := http.NewRequest("GET", urlpath, nil)
	if resumeBool {
		request.Header.Set("Range", fmt.Sprintf("bytes=%d-", fileInfo.Size()))
		offset, _ := fd.Seek(0, io.SeekEnd)
		log.Printf("will resume at %d", offset)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalf("Error client.Do " + err.Error())
	}
	defer response.Body.Close()

	amount := response.ContentLength - fileInfo.Size()
	message := "downloading"
	if resumeBool {
		message = "resuming"
	}
	bar := progressbar.DefaultBytes(amount, message)
	n, err := io.Copy(io.MultiWriter(fd, bar), response.Body)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	bar.Finish()
	log.Printf("Downloaded in output.bin. Success %d\n", n)

	/*
		content, err := io.ReadAll(response.Body)
		if err != nil {
			return false, err
		}

		_, err = output.Write(content)
		if err != nil {
			log.Fatalf("Error output.Write " + err.Error())
		}
	*/
	return true, nil
}

func getSize(urlpath string) string {
	client := http.Client{
		Timeout:   30 * time.Second,
		Transport: http.DefaultTransport,
	}
	response, err := client.Head(urlpath)
	if err != nil {
		log.Fatalf("client.Head error " + err.Error())
	}
	defer response.Body.Close()

	return humanize.Bytes(uint64(response.ContentLength))
}

func getURL(encoded string) string {
	decoded, err := neturl.QueryUnescape(encoded)
	if err != nil {
		return ""
	}

	return decoded
}

func peugeotVersion(VehiculeIdentifier string) (bool, error) {
	if len(VehiculeIdentifier) == 0 {
		return false, fmt.Errorf("Wrong VIN number")
	}

	fmt.Println(fmt.Sprintf("Hello Peugeot from %s", VehiculeIdentifier))

	client := http.Client{
		Timeout:   30 * time.Second,
		Transport: http.DefaultTransport,
	}
	var body = bytes.NewBuffer(createPayload(VehiculeIdentifier))
	var url = UpdateApiURL + fmt.Sprintf("?client_id=%s", ClientID)
	var buf []byte

	response, err := client.Post(
		/*url*/ url,
		/*contentType*/ "application/json",
		/*body*/ body,
	)
	if err != nil {
		log.Fatalf("client.Post error " + err.Error())
	}
	defer response.Body.Close()

	//fmt.Printf("Reponse:\n%+v\n", response)
	buf, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Read response error " + err.Error())
	}

	var result UpdateResult

	err = json.Unmarshal(buf, &result)
	if err != nil {
		log.Fatalf("Unmarshal response error " + err.Error())
	}

	fmt.Printf("Body:\n%+v\n", jsonPrettyPrint(string(buf)))
	//fmt.Printf("Body:\n%+v\n", result)

	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	//for _, soft := range result.Softwares {
	//	fmt.Printf("%s %-20s\n", aliasTypes[soft.SoftwareType], red(soft.CurrentSoftwareVersion))
	//}

	nacSize := getSize(result.Softwares[0].Update[0].UpdateURL)
	rccSize := getSize(result.Softwares[1].Update[0].UpdateURL)
	licenceSize := getSize(result.Softwares[1].Update[0].LicenseURL)
	fmt.Printf("NAC Map\t\t\tCurrent version: %-20s", red(result.Softwares[0].CurrentSoftwareVersion))
	fmt.Printf("\t\tNew version: %-20s\t(Size: %-10s)\n", cyan(result.Softwares[0].Update[0].UpdateVersion), nacSize)
	fmt.Printf("RCC Firwmare\t\tCurrent version: %-20s", red(result.Softwares[1].CurrentSoftwareVersion))
	fmt.Printf("\tNew version: %-20s (Size: %-10s) (License: %-10s)\n", cyan(result.Softwares[1].Update[0].UpdateVersion), rccSize, licenceSize)
	fmt.Printf("Installer URL: %-100s\n", green(getURL(result.InstallerURL)))
	fmt.Printf("Map       URL: %-100s\n", green(getURL(result.Softwares[0].Update[0].UpdateURL)))
	fmt.Printf("Firm      URL: %-100s\n", green(getURL(result.Softwares[1].Update[0].UpdateURL)))
	fmt.Printf("License   URL: %-100s\n", green(getURL(result.Softwares[1].Update[0].LicenseURL)))

	return true, nil
}

func main() {
	var vinFlag = flag.String("vin", "", "help message for (V)ehicule (I)dentifier (N)umber flag")
	flag.Parse()
	if len(*vinFlag) == 0 {
		log.Println("No VIN provided")
		flag.Usage()
		os.Exit(1)
	}

	ok, err := peugeotVersion(*vinFlag)
	if !ok || err != nil {
		log.Fatalf("error " + err.Error())
	}

	os.Exit(0)
}
