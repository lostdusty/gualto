package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/princessmortix/gobalt"
)

var verifyLink = regexp.MustCompile(`[-a-zA-Z0-9@:%_\+.~#?&//=]{2,256}\.[a-z]{2,4}\b(\/[-a-zA-Z0-9@:%_\+.~#?&//=]*)?`)
var count int
var GualtoWin fyne.Window
var onSaveDialog bool

func main() {
	newDownload := gobalt.CreateDefaultSettings()

	gualtoApp := app.NewWithID("com.lostdusty.gualto")
	GualtoWin = gualtoApp.NewWindow("Gualto")
	GualtoWin.CenterOnScreen()
	GualtoWin.Resize(fyne.Size{Width: 800, Height: 400})

	labelMain := widget.NewRichTextFromMarkdown("# Gualto\n\nSave what you love, no extra bullshit. Paste your url below to begin the download.")
	labelMain.Wrapping = fyne.TextWrapWord

	pasteURL := widget.NewEntry()
	pasteURL.PlaceHolder = "Paste your url here..."
	pasteURL.Validator = validation.NewRegexp(`[-a-zA-Z0-9@:%_\+.~#?&//=]{2,256}\.[a-z]{2,4}\b(\/[-a-zA-Z0-9@:%_\+.~#?&//=]*)?`, "Must be a valid link")

	submitURL := widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), nil)
	submitURL.Importance = widget.SuccessImportance
	pasteURL.SetOnValidationChanged(func(err error) {
		if err != nil {
			submitURL.Disable()
		} else {
			submitURL.Enable()
		}
	})

	checkAccordionSettingTwitter := widget.NewCheck("Convert Twitter gifs", func(b bool) {
		newDownload.ConvertTwitterGifs = b
	})
	checkAccordionSettingTwitter.Checked = true

	labelAccordionSettingQuality := widget.NewLabel("Video Quality")
	selAccordionSettingQuality := widget.NewSelect([]string{"360", "480", "720", "1080", "1440", "2160"}, func(s string) {
		newDownload.VideoQuality, _ = strconv.Atoi(s)
	})
	selAccordionSettingQuality.Selected = "1080"
	qualitySettings := container.NewHBox(labelAccordionSettingQuality, selAccordionSettingQuality)

	labelAccordionSettingCodec := widget.NewLabel("Video Codec")
	selAccordionSettingCodec := widget.NewSelect([]string{"h264", "av1", "vp9"}, func(s string) {
		//TODO: Move this to gobalt
		switch s {
		case "h264":
			newDownload.VideoCodec = gobalt.H264
		case "av1":
			newDownload.VideoCodec = gobalt.AV1
		case "vp9":
			newDownload.VideoCodec = gobalt.VP9
		}
	})
	selAccordionSettingCodec.Selected = "h264"
	codecSettings := container.NewHBox(labelAccordionSettingCodec, selAccordionSettingCodec)

	labelAccordionSettingAudio := widget.NewLabel("Audio format")
	selAccordionSettingAudio := widget.NewSelect([]string{"best", "mp3", "ogg", "wav", "opus"}, func(s string) {
		switch s {
		case "best":
			newDownload.AudioCodec = gobalt.Best
		case "mp3":
			newDownload.AudioCodec = gobalt.MP3
		case "ogg":
			newDownload.AudioCodec = gobalt.Ogg
		case "wav":
			newDownload.AudioCodec = gobalt.Wav
		case "opus":
			newDownload.AudioCodec = gobalt.Opus
		}
	})
	selAccordionSettingAudio.SetSelectedIndex(0)
	audioFormatSettings := container.NewHBox(labelAccordionSettingAudio, selAccordionSettingAudio)

	checkAccordionSettingRemoveAudio := widget.NewCheck("Remove audio", func(b bool) {
		newDownload.VideoOnly = b
	})
	checkAccordionSettingRemoveVideo := widget.NewCheck("Remove video", func(b bool) {
		newDownload.AudioOnly = b
	})
	checkAccordionSettingDisableMetadata := widget.NewCheck("Disable metadata", func(b bool) {
		newDownload.DisableVideoMetadata = b
	})
	checkAccordionSettingsFullTikTokAudio := widget.NewCheck("Full tiktok audio", func(b bool) {
		newDownload.FullTikTokAudio = b
	})

	labelFilenamePattern := widget.NewLabel("File name style")
	selFileNamePattern := widget.NewSelect([]string{"classic", "basic", "pretty", "nerdy"}, func(s string) {
		switch s {
		case "basic":
			newDownload.FilenamePattern = gobalt.Basic
		case "classic":
			newDownload.FilenamePattern = gobalt.Classic
		case "nerdy":
			newDownload.FilenamePattern = gobalt.Nerdy
		case "pretty":
			newDownload.FilenamePattern = gobalt.Pretty
		}
	})
	selFileNamePattern.SetSelectedIndex(2)
	filenameSettings := container.NewHBox(labelFilenamePattern, selFileNamePattern)

	accordionMaster := &widget.AccordionItem{
		Title: "Download options",
		Detail: container.NewVBox(checkAccordionSettingTwitter,
			qualitySettings,
			codecSettings,
			audioFormatSettings,
			checkAccordionSettingRemoveAudio,
			checkAccordionSettingRemoveVideo,
			checkAccordionSettingDisableMetadata,
			filenameSettings,
			checkAccordionSettingsFullTikTokAudio,
		),
	}

	accordionOptions := &widget.Accordion{
		Items:     []*widget.AccordionItem{accordionMaster},
		MultiOpen: true,
	}

	/* ABOUt & SETTINGS BUTTONS AT THE END */
	aboutButton := widget.NewButtonWithIcon("about", theme.InfoIcon(), func() {
		infoTextTitle := widget.NewRichTextFromMarkdown("## Gualto")
		infoExit := widget.NewButtonWithIcon("", theme.CancelIcon(), nil)
		infoExit.Importance = widget.DangerImportance
		infoHeader := container.NewBorder(nil, nil, infoTextTitle, infoExit)
		infoText := widget.NewRichTextFromMarkdown("Save what you love, no extra bullshit.\n\nUses [cobalt.tools](cobalt.tools) under the hood.\n\n### Thanks to..\nYou, Wukko, JJ & contributors")
		info := widget.NewModalPopUp(container.NewBorder(infoHeader, nil, nil, nil, infoText), GualtoWin.Canvas())
		infoExit.OnTapped = func() { info.Hide() }
		info.Show()
	})
	aboutButton.Importance = widget.HighImportance
	settingsButton := widget.NewButtonWithIcon("settings", theme.SettingsIcon(), func() {
		storedInstance := gualtoApp.Preferences().StringWithFallback("instance", gobalt.CobaltApi)
		changeInstancesList := &widget.Select{
			Selected: storedInstance,
			Options:  []string{"https://co.wuk.sh", "https://cobalt-api.hyper.lol", "https://coapi.bigbenster702.com", "https://downloadapi.stuff.solutions", "https://cobalt.api.timelessnesses.me", "https://api-dl.cgm.rs", "https://co-api.mae.wtf", "https://capi.oak.li"},
		}
		changeInstancesLabel := widget.NewLabel("This allows you to use a custom instance.\nOnly change if you know what you are doing!")
		changeInstancesList.OnChanged = func(s string) {
			gualtoApp.Preferences().SetString("instance", s)
		}
		settingsDialog := dialog.NewCustom("Gualto App Settings", "Close", container.NewVBox(changeInstancesLabel, changeInstancesList), GualtoWin)
		settingsDialog.Show()
	})
	settingsButton.IconPlacement = widget.ButtonIconTrailingText
	settingsButton.Importance = widget.HighImportance
	submitURL.OnTapped = func() {
		submitURL.Disable()

		newDownload.Url = pasteURL.Text
		downloadMedia(newDownload)
		submitURL.Enable()

	}

	/* CREATE THE FINAL LAYOUT AND DISPLAY */
	submitContainer := container.NewBorder(nil, nil, nil, submitURL, pasteURL)
	downloadActions := container.NewVBox(labelMain, widget.NewSeparator(), submitContainer)
	aboutSettings := container.NewGridWithColumns(2, aboutButton, settingsButton)
	windowContent := container.NewBorder(downloadActions, aboutSettings, nil, nil, container.NewScroll(accordionOptions))
	GualtoWin.SetContent(windowContent)

	gualtoApp.Lifecycle().SetOnEnteredForeground(func() {
		if count == 0 || onSaveDialog {
			count++
			return
		}
		go func() {
			isLink := verifyLink.MatchString(GualtoWin.Clipboard().Content())
			if !isLink {
				return
			}
			downloadClipAsk := dialog.NewConfirm("We found an link!", "Found an url on your clipboard, do you want to paste it?", func(b bool) {
				if b {
					pasteURL.SetText(GualtoWin.Clipboard().Content())
				}
			}, GualtoWin)
			downloadClipAsk.Show()
		}()
	})

	GualtoWin.ShowAndRun()
}

func downloadMedia(options gobalt.Settings) {
	onSaveDialog = true
	statusProgressBar := dialog.NewCustomWithoutButtons("Downloading....", widget.NewProgressBarInfinite(), GualtoWin)
	statusProgressBar.Show()
	cobaltRequestDownloadFile, err := gobalt.Run(options)
	if err != nil {
		dialog.ShowError(err, GualtoWin)
		return
	}
	if cobaltRequestDownloadFile.Status == "picker" {
		dialog.ShowError(fmt.Errorf("Picker is not supported yet."), GualtoWin)
		return
	}

	fmt.Println("Sending request...")
	cobaltMediaResponse, err := http.Get(cobaltRequestDownloadFile.URL)
	if err != nil {
		statusProgressBar.Hide()
		dialog.ShowCustom("error", "close", container.NewVBox(&widget.Label{Text: err.Error(), Wrapping: fyne.TextWrapWord}), GualtoWin)
		return
	}
	fmt.Println("Request send, parsing file...")
	//defer cobaltMediaResponse.Body.Close()
	cobaltGetFileName := cobaltMediaResponse.Header.Get("Content-Disposition")
	_, parseFileName, err := mime.ParseMediaType(cobaltGetFileName)
	mediaFilename := parseFileName["filename"]
	if err != nil {
		mediaFilename = path.Base(cobaltMediaResponse.Request.URL.Path)
	}

	fmt.Println("Filename is:", mediaFilename)

	saveFileDialog := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, GualtoWin)
			return
		}
		if uc == nil {
			return
		}

		fromReqToFile, err := io.Copy(uc, cobaltMediaResponse.Body)
		if err != nil {
			statusProgressBar.Hide()
			dialog.ShowCustom("error", "close", container.NewVBox(&widget.Label{Text: err.Error(), Wrapping: fyne.TextWrapWord}), GualtoWin)
			return
		}
		cobaltMediaResponse.Body.Close()
		statusProgressBar.Hide()
		dialog.ShowInformation(fmt.Sprintf("Downloaded %d.2MB of your media!", (fromReqToFile/1000000)), "The download was sucessful.", GualtoWin)

	}, GualtoWin)
	saveFileDialog.SetFileName(mediaFilename)
	saveFileDialog.Show()

	onSaveDialog = false
}
