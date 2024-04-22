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
	"github.com/lostdusty/gobalt"
)

var verifyLink = regexp.MustCompile(`[-a-zA-Z0-9@:%_\+.~#?&//=]{2,256}\.[a-z]{2,4}\b(\/[-a-zA-Z0-9@:%_\+.~#?&//=]*)?`)
var blockClip bool
var GualtoWin fyne.Window
var instancesList []string

func main() {
	newDownload := gobalt.CreateDefaultSettings() //Create default settings for downloading
	gualtoApp := app.NewWithID("com.lostdusty.gualto")
	GualtoWin = gualtoApp.NewWindow("Gualto")
	GualtoWin.CenterOnScreen()
	GualtoWin.Resize(fyne.Size{Width: 800, Height: 400})

	//Async fetches cobalt instances. If it fails, add only the main instance to the list
	go func() {
		asyncGetCobaltInstances, err := gobalt.GetCobaltInstances()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to fetch more cobalt instances"), GualtoWin)
			instancesList = append(instancesList, gobalt.CobaltApi)
			return
		}

		for _, cobaltInstances := range asyncGetCobaltInstances {
			instancesList = append(instancesList, fmt.Sprintf("https://%v", cobaltInstances.URL))
		}
	}()

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

	checkAccordionSettingTwitter := widget.NewCheck("Don't convert Twitter gifs", func(b bool) {
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
		blockClip = true
		infoTextTitle := widget.NewRichTextFromMarkdown("## Gualto")
		infoExit := widget.NewButtonWithIcon("", theme.CancelIcon(), nil)
		infoExit.Importance = widget.DangerImportance
		infoHeader := container.NewBorder(nil, nil, infoTextTitle, infoExit)
		infoText := widget.NewRichTextFromMarkdown("Save what you love, no extra bullshit.\n\nUses [cobalt.tools](cobalt.tools) under the hood.\n\n### Thanks to..\nYou, Wukko, JJ & contributors")
		info := widget.NewModalPopUp(container.NewBorder(infoHeader, nil, nil, nil, infoText), GualtoWin.Canvas())
		infoExit.OnTapped = func() { info.Hide(); blockClip = false }
		info.Show()
	})
	aboutButton.Importance = widget.HighImportance

	//Settings
	settingsButton := widget.NewButtonWithIcon("settings", theme.SettingsIcon(), func() {
		blockClip = true
		storedInstance := gualtoApp.Preferences().StringWithFallback("instance", gobalt.CobaltApi)
		checkClipboard := gualtoApp.Preferences().BoolWithFallback("clipboard", true)
		changeInstancesList := &widget.Select{
			Selected: storedInstance,
			Options:  instancesList,
		}
		changeInstancesLabel := widget.NewLabel("This allows you to use a custom instance.\nOnly change if you know what you are doing!")
		changeInstancesList.OnChanged = func(s string) {
			gualtoApp.Preferences().SetString("instance", s)
		}
		shouldCheckClipboard := widget.NewCheck("Check clipboard for media to download?", func(b bool) {
			gualtoApp.Preferences().SetBool("clipboard", b)
		})
		shouldCheckClipboard.Checked = checkClipboard
		settingsDialog := dialog.NewCustom("Gualto App Settings", "Close", container.NewVBox(changeInstancesLabel, changeInstancesList, widget.NewSeparator(), shouldCheckClipboard), GualtoWin)
		settingsDialog.Show()
		settingsDialog.SetOnClosed(func() {
			blockClip = false
		})
	})
	settingsButton.IconPlacement = widget.ButtonIconTrailingText
	settingsButton.Importance = widget.HighImportance
	submitURL.OnTapped = func() {
		blockClip = true
		submitURL.Disable()

		statusProgressBar := dialog.NewCustomWithoutButtons("Downloading....", widget.NewProgressBarInfinite(), GualtoWin)
		statusProgressBar.Show()

		newDownload.Url = pasteURL.Text

		err := downloadMedia(newDownload)
		if err != nil {
			errLab := widget.NewLabel(err.Error())
			errLab.Wrapping = fyne.TextWrapWord
			dialog.NewCustom("Somewent went wrong while downloading!", "close", errLab, GualtoWin)
			statusProgressBar.Hide()
		}

		statusProgressBar.Hide()
		submitURL.Enable()
		blockClip = false
	}

	/* CREATE THE FINAL LAYOUT AND DISPLAY */
	submitContainer := container.NewBorder(nil, nil, nil, submitURL, pasteURL)
	downloadActions := container.NewVBox(labelMain, widget.NewSeparator(), submitContainer)
	aboutSettings := container.NewGridWithColumns(2, aboutButton, settingsButton)
	windowContent := container.NewBorder(downloadActions, aboutSettings, nil, nil, container.NewScroll(accordionOptions))
	GualtoWin.SetContent(windowContent)

	gualtoApp.Lifecycle().SetOnEnteredForeground(func() {
		go func() {
			if !blockClip && gualtoApp.Preferences().Bool("clipboard") { //Show clipboard paste if all of these are true
				blockClip = true
				isLink := verifyLink.MatchString(GualtoWin.Clipboard().Content())
				if !isLink {
					return
				}
				downloadClipAsk := dialog.NewConfirm("We found an link!", "Paste URL from clipboard?", func(b bool) {
					if b {
						pasteURL.SetText(GualtoWin.Clipboard().Content())
					}
				}, GualtoWin)
				downloadClipAsk.Show()
				downloadClipAsk.SetOnClosed(func() {
					blockClip = false
				})
			} else {
				return
			}
		}()
	})

	GualtoWin.ShowAndRun()
}

func downloadMedia(options gobalt.Settings) error {
	cobaltRequestDownloadFile, err := gobalt.Run(options)
	if err != nil {
		return err
	}
	if cobaltRequestDownloadFile.Status == "picker" {
		return fmt.Errorf("picker is not supported yet")
	}
	cobaltMediaResponse, err := http.Get(cobaltRequestDownloadFile.URL)
	if err != nil {
		return err
	}
	cobaltGetFileName := cobaltMediaResponse.Header.Get("Content-Disposition")
	_, parseFileName, err := mime.ParseMediaType(cobaltGetFileName)
	mediaFilename := parseFileName["filename"]
	if err != nil {
		mediaFilename = path.Base(cobaltMediaResponse.Request.URL.Path)
	}
	normalizeFileName := regexp.MustCompile(`[^[:word:][:punct:]\s]`)
	normalFileName := normalizeFileName.ReplaceAllString(mediaFilename, "")
	fmt.Printf("old %v, new: %v\n", mediaFilename, normalFileName)

	saveFileDialog := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
		blockClip = true
		if err != nil || uc == nil {
			return
		}
		fromReqToFile, err := io.Copy(uc, cobaltMediaResponse.Body)
		if err != nil {
			dialog.ShowCustom("error", "close", container.NewVBox(&widget.Label{Text: err.Error(), Wrapping: fyne.TextWrapWord}), GualtoWin)
			return
		}
		cobaltMediaResponse.Body.Close()
		dialog.ShowInformation(fmt.Sprintf("Media (%d.2MB) saved with success!", (fromReqToFile/1000000)), fmt.Sprintf("Saved to %v", uc.URI().Path()), GualtoWin)
		blockClip = false
	}, GualtoWin)
	saveFileDialog.SetFileName(normalFileName)
	go saveFileDialog.Show()
	return nil
}
