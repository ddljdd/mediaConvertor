package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func timestamp() string {
	now := time.Now().UTC().Format("15:04:05.99")
	return now
}
func getInfoFile(inputFile string, output *walk.TextEdit) {
	suffIndex := strings.LastIndex(inputFile, ".")
	if suffIndex == -1 {
		output.AppendText("Illegal file name\r\n")
		return
	}
	args := []string{
		"-hide_banner",
		inputFile,
	}
	cmd := exec.Command("ffprobe.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	var cmderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &cmderr
	if err := cmd.Start(); err != nil {
		output.AppendText(err.Error())
		return
	}
	//output.AppendText(timestamp() + " Waiting finished\r\n")
	cmd.Wait()
	output.AppendText(out.String())
	output.AppendText(cmderr.String())
}
func convertFile(inputFile string, output *walk.TextEdit, cfg *Cfg) {
	output.AppendText(timestamp() + " Start to run ffmpeg, format is " + cfg.Format + " file is :" + inputFile + "\r\n")
	suffIndex := strings.LastIndex(inputFile, ".")
	if suffIndex == -1 {
		output.AppendText("Illegal file name\r\n")
		return
	}
	nameIndex := strings.LastIndex(inputFile, "\\")
	outputFile := ""
	if nameIndex == -1 {
		outputFile = "converted_" + inputFile
	} else {
		outputFile = inputFile[:nameIndex+1] + "converted_" + inputFile[nameIndex+1:]
	}

	//outputFile := inputFile + ".new.mkv"
	args := []string{
		"-y", "-hide_banner", "-v", "quiet",
		"-i", inputFile,
	}
	if cfg.Format != "" {
		args = append(args, "-c:v", cfg.Format, "-c:a", "copy")
	}
	if cfg.Rate != "" {
		args = append(args, "-b:v", cfg.Rate)
	}

	if cfg.Fps != "" {
		args = append(args, "-r", cfg.Fps)
	}

	if cfg.Size != "" {
		args = append(args, "-s", cfg.Size)
	}
	args = append(args, outputFile)
	output.AppendText("ffmpeg.exe " + strings.Join(args, " ") + "\r\n")
	cmd := exec.Command("ffmpeg.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	var cmderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &cmderr
	if err := cmd.Start(); err != nil {
		output.AppendText(err.Error())
		return
	}
	output.AppendText(timestamp() + " Waiting finished\r\n")
	cmd.Wait()
	output.AppendText(timestamp() + " finished\r\n")
	//output.AppendText(out.String())
	//output.AppendText(cmderr.String())
}

func startInfo(input *walk.TextEdit, output *walk.TextEdit) {
	//output.AppendText(fmt.Sprintf("%+v", cfg) + "\r\n")
	fileToConvertList := strings.Split(input.Text(), "\r\n")
	for _, f := range fileToConvertList {
		getInfoFile(f, output)
	}
}
func startConvert(input *walk.TextEdit, output *walk.TextEdit, cfg *Cfg) {
	output.AppendText(fmt.Sprintf("%+v", cfg) + "\r\n")
	fileToConvertList := strings.Split(input.Text(), "\r\n")
	for _, f := range fileToConvertList {
		convertFile(f, output, cfg)
	}
}

type Cfg struct {
	Format   string
	Log      string
	Compress int
	Rate     string
	Fps      string
	Size     string
}

func main() {
	var input, output, adv *walk.TextEdit
	var sbi *walk.StatusBarItem
	cfg := new(Cfg)
	cfg.Format = "libx265"
	cfg.Rate = "10M"
	cfg.Log = "Hello world"
	var db *walk.DataBinder
	MainWindow{
		Title:   "Media Convertor",
		MinSize: Size{800, 400},
		Layout:  VBox{},
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     cfg,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
		OnDropFiles: func(files []string) {
			input.SetText(strings.Join(files, "\r\n"))
		},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					TextEdit{AssignTo: &input, Text: "Drag media files to here "},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							PushButton{
								MinSize: Size{60, 40},
								MaxSize: Size{60, 40},
								Text:    "Convert",
								OnClicked: func() {
									db.Submit()
									go startConvert(input, output, cfg)
								},
							},
							PushButton{
								MinSize: Size{60, 40},
								MaxSize: Size{60, 40},
								Text:    "Info",
								OnClicked: func() {
									db.Submit()
									go startInfo(input, output)
								},
							},
						},
					},
				},
			},
			Composite{
				Border:  true,
				MaxSize: Size{400, 80},
				Layout:  Grid{Columns: 6},
				Children: []Widget{
					Label{
						Text: "Convert Format:",
					},
					ComboBox{
						Editable: true,
						Value:    Bind("Format"),
						MaxSize:  Size{80, 0},
						Model:    []string{"libx265", "libx264", "libvpx-vp9"},
					},
					Label{
						Text: "Rate:",
					},
					ComboBox{
						Editable: true,
						Value:    Bind("Rate"),
						MaxSize:  Size{60, 0},
						Model:    []string{"500k", "1M", "2M", "3M", "5M", "10M", "20M", "50M"},
					},
					Label{
						Text: "Advanced option:",
					},
					TextEdit{AssignTo: &adv},
					Label{
						Text: "FPS:",
					},
					ComboBox{
						Editable: true,
						Value:    Bind("Fps"),
						MaxSize:  Size{60, 0},
						Model:    []string{"24", "30", "60"},
					},
					Label{
						Text: "Size:",
					},
					ComboBox{
						Editable: true,
						Value:    Bind("Size"),
						MaxSize:  Size{80, 0},
						Model:    []string{"1920*1080", "1280*720"},
					},
				},
			},
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{

					Label{
						Text: "Compress:",
					},
					Slider{
						Value: Bind("Compress"),
					}},
			},
			TextEdit{AssignTo: &output, VScroll: true},
		},
		StatusBarItems: []StatusBarItem{
			StatusBarItem{
				AssignTo: &sbi,
				Text:     "click",
				Width:    80,
				OnClicked: func() {
					if sbi.Text() == "click" {
						sbi.SetText("again")
					} else {
						sbi.SetText("click")
					}
				},
			},
			StatusBarItem{
				Text:        "left",
				ToolTipText: "no tooltip for me",
			},
		},
	}.Run()
}
