package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"
	"github.com/tidwall/gjson"
)

const BaseURL = "http://localhost:8080/api/v1"

var client *resty.Client

func main() {
	client = resty.New()
	client.SetBaseURL(BaseURL)
	// Analysis can take a long time, so we set a long timeout
	client.SetTimeout(3 * time.Minute)

	// Header
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("MaiEcho CLI Tester")
	pterm.Info.Println("Connecting to " + BaseURL)

	// Check Status
	checkSystemStatus()

	// Main Loop
	for {
		options := []string{
			"1. 🔍 搜索歌曲 (Search Songs)",
			"2. 📄 歌曲详情 (Get Song Details)",
			"3. 🕷️ 触发采集 (Trigger Collection)",
			"4. 🧠 触发分析 (Trigger Analysis)",
			"5. 📊 查看分析结果 (Get Analysis Result)",
			"6. 🔄 同步数据 (Sync Data)",
			"7. 🏷️ 别名查询 (Alias Query)",
			"8. 🚪 退出 (Exit)",
		}

		printer := pterm.DefaultInteractiveSelect.WithOptions(options)
		printer.DefaultText = "请选择操作"
		selectedOption, _ := printer.Show()

		switch selectedOption {
		case options[0]:
			searchSongs()
		case options[1]:
			getSongDetails()
		case options[2]:
			triggerCollection()
		case options[3]:
			triggerAnalysis()
		case options[4]:
			getAnalysisResult()
		case options[5]:
			syncData()
		case options[6]:
			searchAliases()
		case options[7]:
			pterm.Info.Println("Bye!")
			return
		}
		fmt.Println()
	}
}

func checkSystemStatus() {
	spinner, _ := pterm.DefaultSpinner.Start("Checking system status...")
	resp, err := client.R().Get("/system/status")
	if err != nil {
		spinner.Fail("Connection failed: " + err.Error())
		return
	}
	if resp.IsError() {
		spinner.Fail("Server returned error: " + resp.Status())
		return
	}
	spinner.Success("System Online: " + string(resp.Body()))
}

func searchSongs() {
	prompt := pterm.DefaultInteractiveTextInput.WithDefaultText("输入关键词 (Keyword)")
	keyword, _ := prompt.Show()

	spinner, _ := pterm.DefaultSpinner.Start("Searching...")
	resp, err := client.R().SetQueryParam("keyword", keyword).Get("/songs")
	if err != nil {
		spinner.Fail(err.Error())
		return
	}
	spinner.Success("Search completed")

	data := gjson.GetBytes(resp.Body(), "items")
	if len(data.Array()) == 0 {
		pterm.Warning.Println("No songs found.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Title", "Artist", "Type"})

	data.ForEach(func(key, value gjson.Result) bool {
		table.Append([]string{
			value.Get("id").String(),
			value.Get("title").String(),
			value.Get("artist").String(),
			value.Get("type").String(),
		})
		return true
	})
	table.Render()
}

func getSongDetails() {
	id := askForID("输入歌曲 ID (Game ID)")
	if id == "" {
		return
	}

	spinner, _ := pterm.DefaultSpinner.Start("Fetching details...")
	resp, err := client.R().Get("/songs/" + id)
	if err != nil {
		spinner.Fail(err.Error())
		return
	}
	if resp.IsError() {
		spinner.Fail("Error: " + resp.Status())
		return
	}
	spinner.Success("Details fetched")

	json := gjson.ParseBytes(resp.Body())
	pterm.DefaultSection.Println(json.Get("title").String())
	pterm.Info.Println("Artist: " + json.Get("artist").String())
	pterm.Info.Println("BPM: " + json.Get("bpm").String())

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Difficulty", "Level", "DS", "Notes"})

	json.Get("charts").ForEach(func(key, value gjson.Result) bool {
		table.Append([]string{
			value.Get("difficulty").String(),
			value.Get("level").String(),
			value.Get("ds").String(),
			value.Get("notes").String(),
		})
		return true
	})
	table.Render()
}

func triggerCollection() {
	id := askForID("输入歌曲 ID (Game ID)")
	if id == "" {
		return
	}

	intID, _ := strconv.Atoi(id)
	spinner, _ := pterm.DefaultSpinner.Start("Triggering collection...")
	resp, err := client.R().SetBody(map[string]interface{}{
		"game_id": intID,
	}).Post("/collect")

	if err != nil {
		spinner.Fail(err.Error())
		return
	}
	if resp.IsError() {
		spinner.Fail("Error: " + resp.Status())
		return
	}
	spinner.Success("Collection triggered successfully!")
}

func triggerAnalysis() {
	id := askForID("输入歌曲 ID (Game ID)")
	if id == "" {
		return
	}

	spinner, _ := pterm.DefaultSpinner.Start("Triggering analysis (Async)...")
	resp, err := client.R().Post("/analysis/songs/" + id)

	if err != nil {
		spinner.Fail(err.Error())
		return
	}
	if resp.IsError() {
		spinner.Fail("Error: " + resp.Status())
		return
	}
	spinner.Success("Analysis task started!")
}

func getAnalysisResult() {
	id := askForID("输入歌曲 ID (Game ID)")
	if id == "" {
		return
	}

	spinner, _ := pterm.DefaultSpinner.Start("Fetching analysis result...")
	resp, err := client.R().Get("/analysis/songs/" + id)

	if err != nil {
		spinner.Fail(err.Error())
		return
	}
	if resp.IsError() {
		spinner.Fail("Error: " + resp.Status())
		return
	}
	spinner.Success("Result fetched")

	result := gjson.ParseBytes(resp.Body())

	// Song Summary
	songResult := result.Get("song_result")
	if songResult.Exists() {
		pterm.DefaultSection.Println("🎵 Song Analysis: " + songResult.Get("summary").String())
		pterm.DefaultBox.WithTitle("Rating Advice").Println(songResult.Get("rating_advice").String())
	} else {
		pterm.Warning.Println("No song-level analysis found.")
	}

	// Chart Results
	chartResults := result.Get("chart_results")
	if chartResults.Exists() && len(chartResults.Array()) > 0 {
		pterm.DefaultSection.Println("📈 Chart Analysis Details")
		chartResults.ForEach(func(key, value gjson.Result) bool {
			diff := value.Get("difficulty").String()

			pterm.DefaultPanel.WithPanels(pterm.Panels{
				{{Data: pterm.NewStyle(pterm.FgYellow).Sprintf("Target ID: %s [%s]", value.Get("target_id").String(), diff)}},
				{{Data: value.Get("summary").String()}},
				{{Data: pterm.NewStyle(pterm.FgCyan).Sprint("Difficulty Analysis:\n") + value.Get("difficulty_analysis").String()}},
				{{Data: pterm.NewStyle(pterm.FgGreen).Sprint("Rating Advice:\n") + value.Get("rating_advice").String()}},
			}).Render()
			fmt.Println()
			return true
		})
	} else {
		pterm.Info.Println("No chart-specific analysis available yet.")
	}
}

func syncData() {
	spinner, _ := pterm.DefaultSpinner.Start("Syncing data from Diving-Fish...")
	resp, err := client.R().Post("/songs/sync")
	if err != nil {
		spinner.Fail(err.Error())
		return
	}
	if resp.IsError() {
		spinner.Fail("Error: " + resp.Status())
		return
	}
	spinner.Success("Sync command sent.")
}

func searchAliases() {
	prompt := pterm.DefaultInteractiveTextInput.WithDefaultText("输入别名关键词 (Alias Keyword)")
	keyword, _ := prompt.Show()

	spinner, _ := pterm.DefaultSpinner.Start("Searching aliases...")
	resp, err := client.R().SetQueryParam("keyword", keyword).Get("/songs")
	if err != nil {
		spinner.Fail(err.Error())
		return
	}
	spinner.Success("Search completed")

	data := gjson.GetBytes(resp.Body(), "items")
	if len(data.Array()) == 0 {
		pterm.Warning.Println("No songs found matching the alias.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Title", "Aliases"})
	table.SetAutoWrapText(false)

	data.ForEach(func(key, value gjson.Result) bool {
		var aliases []string
		value.Get("aliases").ForEach(func(k, v gjson.Result) bool {
			aliases = append(aliases, v.Get("alias").String())
			return true
		})

		aliasStr := ""
		if len(aliases) > 0 {
			if len(aliases) > 5 {
				aliasStr = fmt.Sprintf("%s, ... (+%d)", strings.Join(aliases[:5], ", "), len(aliases)-5)
			} else {
				aliasStr = strings.Join(aliases, ", ")
			}
		}

		table.Append([]string{
			value.Get("id").String(),
			value.Get("title").String(),
			aliasStr,
		})
		return true
	})
	table.Render()
}

func askForID(text string) string {
	prompt := pterm.DefaultInteractiveTextInput.WithDefaultText(text)
	result, _ := prompt.Show()
	return result
}
