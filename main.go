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
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lostdusty/gobalt"
)

var (
	verifyLink      = regexp.MustCompile(`[-a-zA-Z0-9@:%_\+.~#?&//=]{2,256}\.[a-z]{2,4}\b(\/[-a-zA-Z0-9@:%_\+.~#?&//=]*)?`)
	blockClip       bool
	GualtoWin       fyne.Window
	GualtoApp       fyne.App
	cobaltInstances []string
)

func discoverCobaltInstances() {
	asyncGetCobaltInstances, err := gobalt.GetCobaltInstances()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to fetch more cobalt instances"), GualtoWin)
		cobaltInstances = append(cobaltInstances, gobalt.CobaltApi)
		return
	}

	for _, dcobaltInstances := range asyncGetCobaltInstances {
		cobaltInstances = append(cobaltInstances, fmt.Sprintf("https://%v", dcobaltInstances.URL))
	}

}

func main() {
	newDownload := gobalt.CreateDefaultSettings() //Create default settings for downloading
	GualtoApp = app.NewWithID("com.lostdusty.gualto")
	GualtoWin = GualtoApp.NewWindow("Gualto")
	GualtoWin.CenterOnScreen()
	GualtoWin.Resize(fyne.Size{Width: 600, Height: 400})

	/* APP SETTINGS GETTERS
	 */
	storedInstance := GualtoApp.Preferences().StringWithFallback("instance", gobalt.CobaltApi)
	checkClipboard := GualtoApp.Preferences().BoolWithFallback("clipboard", true)
	shouldRememberPath := GualtoApp.Preferences().BoolWithFallback("remember-path", true)
	GualtoApp.Preferences().StringWithFallback("path", "")
	newTheme := GualtoApp.Preferences().BoolWithFallback("theme", false)
	if newTheme {
		GualtoApp.Settings().SetTheme(cobaltTheme{})
	}
	/* END OF THE APP SETTINGS SECTION
	 */

	//Async fetches cobalt instances. If it fails, add only the main instance to the list
	go discoverCobaltInstances()

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

	/* Author: LD
	PART I: Ui components for the download settings related children.

	SECTION I: GENERAL OPTIONS
	MODIFIED: 03/07/2024
	*/
	downloadSettingLabelGeneral := widget.NewRichTextFromMarkdown("### General")
	// "General"

	downloadSettingDisableMetadata := widget.NewCheck("Disable metadata?", func(b bool) {
		newDownload.DisableVideoMetadata = b
	})
	//[] Disable metadata?

	downloadSettingLabelFilenamePattern := widget.NewLabel("Name file as:")
	downloadSettingSelectFilenamePattern := widget.NewSelect([]string{"classic", "basic", "pretty", "nerdy"}, func(s string) {
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
	downloadSettingSelectFilenamePattern.SetSelectedIndex(0)
	//Name file as: [Basic]

	boxFilenameSettings := container.NewHBox(downloadSettingLabelFilenamePattern, downloadSettingSelectFilenamePattern)
	// Container to group checkbox & text to make: [] Disable metadata?

	groupGeneralDownload := container.NewVBox(downloadSettingLabelGeneral, downloadSettingDisableMetadata, boxFilenameSettings)
	//Merge them all like:
	// ## General
	// [] Disable metadata?
	// Name file as: [Basic]

	/*	SECTION II: VIDEO SETTINGS	*/

	downloadSettingLabelVideo := widget.NewRichTextFromMarkdown("### Video")
	// ## Video

	downloadSettingLabelQuality := widget.NewLabel("Video Quality:")
	downloadSettingSelectQuality := widget.NewSelect([]string{"144", "240", "360", "480", "720", "1080", "1440", "2160"}, func(s string) {
		newDownload.VideoQuality, _ = strconv.Atoi(s)
	})
	downloadSettingSelectQuality.Selected = "1080"
	boxQualitySettings := container.NewHBox(downloadSettingLabelQuality, downloadSettingSelectQuality)
	// Video Quality: [1080]

	downloadSettingLabelVideoCodec := widget.NewLabel("Youtube Video Codec:")
	downloadSettingSelectVideoCodec := widget.NewSelect([]string{"h264", "av1", "vp9"}, func(s string) {
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
	downloadSettingSelectVideoCodec.Selected = "h264"
	boxCodecSettings := container.NewHBox(downloadSettingLabelVideoCodec, downloadSettingSelectVideoCodec)
	// Youtube Video Codec: [h264]

	downloadSettingRemoveVideo := widget.NewCheck("Remove video?", func(b bool) {
		newDownload.AudioOnly = b
	})
	// [] Remove video?

	groupVideoDownload := container.NewVBox(downloadSettingLabelVideo, boxQualitySettings, boxCodecSettings, downloadSettingRemoveVideo)
	//Merge them like:
	// ## Video
	// Video Quality: [1080]
	// Youtube Video Codec: [h264]
	// [] Remove video?

	/*	SECTION III: AUDIO	*/
	downloadSettingLabelAudio := widget.NewRichTextFromMarkdown("### Audio")
	// ## Audio

	downloadSettingLabelAudioCodec := widget.NewLabel("Audio format:")
	downloadSettingSelectAudioCodec := widget.NewSelect([]string{"best", "mp3", "ogg", "wav", "opus"}, func(s string) {
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
	downloadSettingSelectAudioCodec.SetSelectedIndex(0)
	boxAudioSettings := container.NewHBox(downloadSettingLabelAudioCodec, downloadSettingSelectAudioCodec)
	// Audio format: [best]

	downloadSettingRemoveAudio := widget.NewCheck("Remove audio?", func(b bool) {
		newDownload.VideoOnly = b
	})
	// [] Remove audio?

	groupAudioDownload := container.NewVBox(downloadSettingLabelAudio, boxAudioSettings, downloadSettingRemoveAudio)
	//Merge them like:
	// ## Audio
	// Audio format: [best]
	// [] Remove audio?

	/*	SECTION IV: PLATFORM SPECIFIC SETTINGS	*/
	downloadSettingLabelPlatform := widget.NewRichTextFromMarkdown("### Platform specific")
	// ## Platform specific

	downloadSettingTwitter := widget.NewCheck("Convert Tweets to gifs?", func(b bool) {
		newDownload.ConvertTwitterGifs = b
	})
	downloadSettingTwitter.Checked = true
	// [X] Convert Tweets to gifs?

	downloadSettingTiktok := widget.NewCheck("Full tiktok audio?", func(b bool) {
		newDownload.FullTikTokAudio = b
	})
	// [] Full tiktok audio?

	groupPlatformSpecific := container.NewVBox(downloadSettingLabelPlatform, downloadSettingTwitter, downloadSettingTiktok)
	//Merge them like:
	// ## Platform specific
	// [X] Convert Tweets to gifs?
	// [] Full tiktok audio?

	/*	SECTION V: MERGE INTO A SINGLE CONTAINER	*/
	leftOptions := container.NewVBox(groupGeneralDownload, groupVideoDownload)
	rightOptions := container.NewVBox(groupAudioDownload, groupPlatformSpecific)
	//sep := canvas.NewLine(theme.PrimaryColor())
	downloadOptions := container.NewAdaptiveGrid(2, leftOptions, rightOptions)
	/*
	* END OF PART I.
	 */

	/* Author: LD
	PART II: Layout for the tab "about".

	SECTION I: Text Definition
	MODIFIED: 04/07/2024
	*/
	aboutTabMainText := widget.NewRichTextFromMarkdown("# About\n\nSave what you love, no extra bullshit.\n\nUses [cobalt.tools](cobalt.tools) under the hood.\n\n### Thanks to..\nYou, Wukko, JJ & contributors")
	aboutTab := container.NewTabItemWithIcon("", theme.InfoIcon(), aboutTabMainText)
	/*
	* END OF PART II.
	 */

	/* Author: LD
	PART III: Layout for the tab "settings".

	SECTION I: GET SETTINGS
	MODIFIED: 04/07/2024
	*/
	tabSettingsSelectInstance := &widget.Select{
		Selected: storedInstance,
		Options:  cobaltInstances,
	}

	/*	SECTION II: CREATE LAYOUT TEXT + INSTANCE CHANGER	*/
	tabSettingsLabelInstance := widget.NewRichTextFromMarkdown("# Settings\n\nThis settings allows you to use a custom instance.\n\nYou might want to change this if you're getting any issues with the selected instance.")
	tabSettingsSelectInstance.OnChanged = func(s string) {
		GualtoApp.Preferences().SetString("instance", s)
	}

	/*	SECTION III: OPTION TO SCAN CLIPBOARD FOR DOWNLOADABLE LINKS	*/
	tabSettingsCheckClip := widget.NewCheck("Check clipboard for links to download?", func(b bool) {
		GualtoApp.Preferences().SetBool("clipboard", b)
	})
	tabSettingsCheckClip.Checked = checkClipboard

	/*	SECTION IV: CUSTOM COBALT THEME?	*/
	tabSettingsCustomTheme := widget.NewCheck("Use new app theme?", func(b bool) {
		GualtoApp.Preferences().SetBool("theme", b)
		if b {
			GualtoApp.Settings().SetTheme(cobaltTheme{})
		} else {
			GualtoApp.Settings().SetTheme(theme.DefaultTheme())
		}
	})
	tabSettingsCustomTheme.Checked = newTheme

	/*	SECTION V: REMEMBER LAST PATH WHERE IT WAS DOWNLOADED?	*/
	/*	DESKTOP ONLY: ANDROID & iOS PICKER ALREADY DOES THAT	*/
	tabSettingsLastPath := widget.NewCheck("Remember last folder saved?", func(b bool) {
		GualtoApp.Preferences().SetBool("remember-path", b)
	})
	tabSettingsLastPath.Checked = shouldRememberPath
	if fyne.CurrentApp().Driver().Device().IsMobile() {
		tabSettingsLastPath.Hide()
	}

	tabSettingsContent := container.NewVBox(tabSettingsLabelInstance, tabSettingsSelectInstance, widget.NewSeparator(), tabSettingsCheckClip, widget.NewSeparator(), tabSettingsCustomTheme, widget.NewSeparator(), tabSettingsLastPath)
	settingsTab := container.NewTabItemWithIcon("", theme.SettingsIcon(), tabSettingsContent)

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
	downloadActions := container.NewVBox(labelMain, submitContainer)
	tabMainContent := container.NewBorder(downloadActions, nil, nil, nil, container.NewScroll(downloadOptions))
	mainTab := container.NewTabItemWithIcon("", theme.HomeIcon(), tabMainContent)

	layoutTabs := container.NewAppTabs(mainTab, settingsTab, aboutTab)
	layoutTabs.OnSelected = func(ti *container.TabItem) {
		if layoutTabs.SelectedIndex() == 1 && len(cobaltInstances) > 1 {
			tabSettingsSelectInstance.SetOptions(cobaltInstances) //Fix to set cobalt instances, for some reason it's not being set anymore.
		}
	}
	layoutTabs.SetTabLocation(container.TabLocationLeading)

	GualtoWin.SetContent(layoutTabs)

	GualtoApp.Lifecycle().SetOnEnteredForeground(func() {

		go func() {
			if !blockClip && GualtoApp.Preferences().Bool("clipboard") { //Show clipboard paste if all of these are true
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

	saveFileDialog := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
		savingFile := dialog.NewProgressInfinite("Downloading your file...", "might take a while.", GualtoWin)
		savingFile.Show()
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
		if GualtoApp.Preferences().Bool("remember-path") {
			GualtoApp.Preferences().SetString("path", uc.URI().Path())
		}
		savingFile.Hide()
		dialog.ShowInformation(fmt.Sprintf("Media (%d.2MB) saved with success!", (fromReqToFile/1000000)), fmt.Sprintf("Saved to %v", uc.URI().Path()), GualtoWin)
		blockClip = false
	}, GualtoWin)
	if GualtoApp.Preferences().String("path") != "" {
		u, _ := storage.ParseURI(GualtoApp.Preferences().String("path"))
		ul, _ := storage.ListerForURI(u)
		saveFileDialog.SetLocation(ul)
	}
	saveFileDialog.SetFileName(normalFileName)
	go saveFileDialog.Show()
	return nil
}
