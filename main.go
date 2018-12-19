package main

import (
	"fmt"
	"encoding/hex"
	"log"
	"encoding/binary"
	"math"
	"encoding/json"
	"github.com/tarm/serial"
	"time"

	"net/http"
)


/*
set GOARCH=amd64
set GOOS=windows
go build -o mtdugraber
go run mtdu_graber.go


for beagleBone
set GOARCH=arm
set GOOS=linux
go build -o mtdugraber
*/
//стурктура для хранения настроек
type Options struct {
	Address string
	Port string

	AdminLogin string
	AdminPass string

	DeviceI2C string
	DeviceVideo string
	DeviceModem string

	APN string
	PIN string
	User string
	Password string

	VPNType int

	AddressTunnel string
	UserTunnel string
	PasswordTunnel string

	ZeroName string

	LDebug int
	FDebug bool //флаг отладки для логов
}

var (
	logTypeMess = []string{"blank", "error", "info", "verbose"}

	//devices[] DevI2C = nil

	//все изменяемые опции
	options = Options {
		Address: "0.0.0.0",
		Port: "8080",
		AdminLogin: "aplit",
		AdminPass: "aplit",
		DeviceI2C: "/dev/i2c-1",
		DeviceVideo: "/dev/video0",
		DeviceModem: "/dev/ttyUSB2",

		APN: "internet",
		PIN: "",
		User: "user",
		Password: "pass",

		//VPNType: VPN_PPTP,

		AddressTunnel: "109.226.44.112",
		UserTunnel: "dragon1",
		PasswordTunnel: "dragon12",

		ZeroName: "9f77fc393efb6456",

	//	LDebug: MESS_VERBOSE,
		FDebug: true,
	}

	//функции для обработки web api
	/*
	processingWeb = []ProcessingWeb{
		{"getValue", processApiGetValue},
		{"checkInet", processApiCheckInet},
		{"getOptions", processApiGetOptions},
		{"setOptions", processApiSetOptions},
		{"reloadDevice", processApiReloadDevice},
		{"reloadModem", processApiReloadModem},
		{"getLog", processApiGetLog} }

	attemptModemRestart = 0

	logFile *os.File*/
)

type DataPeriph struct {
	NoteUseField 		uint16
	ValOut_U 			float32
	ValOut_I 			float32
	Rezerv 				float32
	Status 				uint16
	UstOutManualMode_I 	float32
	ValMaxOut_I 		float32
	ValTemperature_one 	float32
	ValTemperature_two 	float32
	ValUSety 			float32
	ValRLoad 			float32
	UstOutAutoMode_I 	float32
	RangeOutMin_I 		float32
	RangeOutMax_I 		float32
	RangeOutMin_U 		float32
	RangeOutMax_U 		float32
	EnableWork 			uint16
	DeviceAddr 			uint8
}

type ReqSetDataPeriph struct {
	UstOutAutoMode_I 	float32
	RangeOutMin_I 		float32
	RangeOutMax_I 		float32
	RangeOutMin_U 		float32
	RangeOutMax_U 		float32
	EnableWork 			uint16
	DeviceAddr 			uint8
}


type DataPeriphROM struct {
	ValOut_U 			float32
	ValOut_I 			float32
	Val_I1	 			float32
	Val_I2	 			float32
	Status 				uint16
	UstOutManualMode_I 	float32
	ValMaxOut_I 		float32
	ValTemperature_one 	float32
	ValTemperature_two 	float32
	ValUSety 			float32
	ValRLoad 			float32
}


func main(){
	//fmt.Println("dd")
	fmt.Println("Данные с МТДУ")
	// Open wikipedia in a 800x600 resizable window

	//	01 10 0016 000B 16 3F000000 00000000 42200000 42C80000 42DC0000 0001	8E
	//	01100016000B163F000000000000004220000042C8000042DC000000018E
	//const s = "04034038AE147C3EB40A3ED6E6B70C394043E900CB0000000042480000C2480000C2480000435BCE270000000000000000000000004220000042C8000042DC0000000179"

	//Данные через GPRS
	//"00 03 16 410090BA41266B9B80294EBF897B8BE100CB0000000052"
	const s = "000316411E7A803DC415C2506E84BD30E0FCC000CB0000000047"
	decoded, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Len = ",len(decoded))
	//fmt.Println("decoded = ",decoded)
	periphROM := DataPeriphROM{}
	buffToDataPeriphROM(decoded,&periphROM)

	//bbb,_:= json.Marshal(periphROM)
	//fmt.Println(string(bbb))

	//return

//	bits := binary.BigEndian.Uint32([]byte{0x38,0xAE,0x14,0x7C})
//	bfloat := math.Float32frombits(bits)
//	fmt.Println(bfloat)

	//fmt.Println("len buff in hex = ", len(decoded))
	periph := DataPeriph{}
	req_periph := ReqSetDataPeriph{}
	buffToDataPeriph(decoded,&periph)
	//fmt.Printf("dataperiph", periph)
	//b,_:= json.Marshal(periph)
	//fmt.Println(string(b))

	c := &serial.Config{Name: "COM3", Baud: 9600,ReadTimeout:time.Millisecond*5000}
	//c := &serial.Config{Name: "/dev/ttyO1", Baud: 9600,ReadTimeout:time.Millisecond*5000}
	sp, err1 := serial.OpenPort(c)

	if err1 != nil {
		fmt.Println("Not open COM3")
		log.Fatal(err1)
	}


	//go httpServer()

	// websocket server
	server := NewServer("/ws")
	go server.Listen()
	//http.Handle("/", http.FileServer(http.Dir("webroot")))



	go func() {
		for{
			buff := readFromSerial(sp)
			decoded,err := hex.DecodeString(string(buff))
			if(err != nil){
				fmt.Println(err)
				log.Println(err)
				log.Println("Error from buff")
				log.Println(string(buff))
				//fmt.Println("Error from buff")
				fmt.Println(string(buff))
			}else {
				if((len(buff)>120)&&((len(buff)<200))){

					buffToDataPeriph(decoded,&periph)
					b,_:= json.Marshal(periph)
					fmt.Println(string(b))
					str := string(b)
					server.SendAll(&str)
				}else{
					if(len(buff)>50){

						buffToReqSetDataPeriph(decoded,&req_periph)
						b,_:= json.Marshal(req_periph)
						fmt.Println(string(b))
					}else {
						if(len(buff)>6){
							addr:= decoded[0]
							fmt.Println("Req GET_REG from addr ",addr)
						}else {
							log.Println("NOT DETECT REQ")
							//fmt.Println("NOT DETECT REQ")
							fmt.Println(string(buff))
						}

					}
				}
			}

		}

	}()


	//go http.ListenAndServe("192.168.1.33:8080", nil)
	go http.ListenAndServe("127.0.0.1:8080", nil)
	/*
	err = http.ListenAndServe("127.0.0.1:8080", nil)
	if err != nil {
		log.Println("webServer couldn't take port: " )
		//addLog(MESS_ERROR, "webServer couldn't take port: " + fmt.Sprint(err))
	}
	*/


//time.Sleep(time.Second*5)
	var input string
	//str := "Hi from socket"
	for input != "q" {
		fmt.Scanln(&input)

		//server.SendAll(&str)
	}
/*
	c := &serial.Config{Name: "COM3", Baud: 9600,ReadTimeout:time.Millisecond*5000}
	s, err := serial.OpenPort(c)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	buf := make([]byte, 128)
	n, err := s.Read(buf)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	fmt.Println("n = ",n)
	fmt.Println("buf = ",buf[:n])
	log.Printf("%q", buf[:n])

	i:=0
	buff_byte := make([]byte, 0)
	//buff_byte
	for (i < 100) {
		s.Read(buff_byte)
		fmt.Println(buff_byte[0]-0x30)
		i = i+1
	}*/
}
func readFromSerial(s *serial.Port)[]byte  {
	f_end:= false


	buff_byte := make([]byte, 1)
	ret_buff := make([]byte, 0)
	for !f_end {
		s.Read(buff_byte)
		if(buff_byte[0] == ':'){
			ret_buff = nil
			//ret_buff.
		}else{
			if(buff_byte[0] == 0x0d){
				f_end = true
			}else {
				ret_buff = append(ret_buff, buff_byte[0])
			}
		}
		//f_end = true
	}
	return ret_buff
}




func buffToDataPeriphROM(data []byte, periph *DataPeriphROM) {

	//Данные через GPRS
	//"00 03 16 410090BA	41266B9B80294EBF897B8BE100CB0000000052"
	/*
	ValOut_U 			float32
	ValOut_I 			float32
	Val_I1	 			float32
	Val_I2	 			float32
	Status 				uint16
	UstOutManualMode_I 	float32
	ValMaxOut_I 		float32
	ValTemperature_one 	float32
	ValTemperature_two 	float32
	ValUSety 			float32
	ValRLoad 			float32
	*/
	if(len(data)<22){
		return
	}


	bits := binary.BigEndian.Uint32(data[3:7])
	periph.ValOut_U = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[7:7+4])
	periph.ValOut_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[11:11+4])
	periph.Val_I1 = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[15:15+4])
	periph.Val_I2 = math.Float32frombits(bits)
	periph.Status = (uint16(data[19]) <<8)|uint16(data[20])

	bits = binary.BigEndian.Uint32(data[21:21+4])
	periph.UstOutManualMode_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[25:25+4])
	periph.ValMaxOut_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[29:29+4])
	periph.ValTemperature_one = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[33:33+4])
	periph.ValTemperature_two = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[37:37+4])
	periph.ValUSety = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[41:41+4])
	periph.ValRLoad = math.Float32frombits(bits)

}



func buffToDataPeriph(data []byte, periph *DataPeriph) {

	if(len(data)<60){
		return
	}
	periph.DeviceAddr = data[0]
	periph.NoteUseField = (uint16(data[1]) <<8)|uint16(data[2])
	bits := binary.BigEndian.Uint32(data[3:7])
	periph.ValOut_U = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[7:7+4])
	periph.ValOut_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[11:11+4])
	periph.Rezerv = math.Float32frombits(bits)

	periph.Status = (uint16(data[19]) <<8)|uint16(data[20])
	bits = binary.BigEndian.Uint32(data[21:21+4])
	periph.UstOutManualMode_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[25:25+4])
	periph.ValMaxOut_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[29:29+4])
	periph.ValTemperature_one = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[33:33+4])
	periph.ValTemperature_two = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[37:37+4])
	periph.ValUSety = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[41:41+4])
	periph.ValRLoad = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[45:45+4])
	periph.UstOutAutoMode_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[49:49+4])
	periph.RangeOutMin_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[53:53+4])
	periph.RangeOutMax_I = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[57:57+4])
	periph.RangeOutMin_U = math.Float32frombits(bits)
	bits = binary.BigEndian.Uint32(data[61:61+4])
	periph.RangeOutMax_U = math.Float32frombits(bits)

	periph.EnableWork = (uint16(data[65]) <<8)|uint16(data[66])

}

func buffToReqSetDataPeriph(data []byte, reqdata *ReqSetDataPeriph) {
	//	01 10 0016 000B 16 3F000000 00000000 42200000 42C80000 42DC0000 0001	8E
	//	01100016000B163F000000000000004220000042C8000042DC000000018E
	if(len(data) < 30){
		return
	}
	reqdata.DeviceAddr = data[0]
	n_func := data[1]
	if(n_func == 0x10){
		//bits := binary.BigEndian.Uint32(data[3:7])
		bits := binary.BigEndian.Uint32(data[7:7+4])
		reqdata.UstOutAutoMode_I = math.Float32frombits(bits)
		bits = binary.BigEndian.Uint32(data[11:11+4])
		reqdata.RangeOutMin_I = math.Float32frombits(bits)
		bits = binary.BigEndian.Uint32(data[15:15+4])
		reqdata.RangeOutMax_I = math.Float32frombits(bits)
		bits = binary.BigEndian.Uint32(data[19:19+4])
		reqdata.RangeOutMin_U = math.Float32frombits(bits)
		bits = binary.BigEndian.Uint32(data[23:23+4])
		reqdata.RangeOutMax_U = math.Float32frombits(bits)

		reqdata.EnableWork = (uint16(data[27]) <<8)|uint16(data[28])

	}else{
		return
	}
}

