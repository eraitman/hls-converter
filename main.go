package main

import (
	"context"
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type item struct {
	path     string
	state    string // 대기 / 변환 중 / 완료 / 실패
	progress float64
	detail   string
}

func main() {
	a := app.New()
	a.Settings().SetTheme(modernTheme{})
	w := a.NewWindow("HLS 변환기")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 창을 닫으면 진행 중인 ffmpeg도 종료
	buildUI(ctx, w)
	w.Resize(fyne.NewSize(640, 480))
	w.ShowAndRun()
}

func card(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(colCard)
	bg.CornerRadius = 10
	bg.StrokeColor = colBorder
	bg.StrokeWidth = 1
	return container.NewStack(bg, container.NewPadded(content))
}

func buildUI(ctx context.Context, w fyne.Window) {
	var items []*item
	running := false

	// 헤더
	title := canvas.NewText("HLS 변환기", colText)
	title.TextSize = 22
	title.TextStyle.Bold = true
	subtitle := canvas.NewText("mp4 → 480p · 720p · 1080p HLS (libx264)", colMuted)
	subtitle.TextSize = 12
	header := container.NewVBox(title, subtitle)

	// 빈 상태: 드롭존
	dropIcon := widget.NewIcon(theme.UploadIcon())
	dropText := canvas.NewText("mp4 파일을 여기에 끌어다 놓으세요", colText)
	dropText.TextSize = 16
	dropText.Alignment = fyne.TextAlignCenter
	dropHint := canvas.NewText("여러 개를 한 번에 놓으면 순서대로 변환합니다", colMuted)
	dropHint.TextSize = 12
	dropHint.Alignment = fyne.TextAlignCenter
	iconBox := container.NewGridWrap(fyne.NewSize(56, 56), dropIcon)
	dropZone := card(container.NewCenter(container.NewVBox(container.NewCenter(iconBox), dropText, dropHint)))

	// 파일 목록
	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject {
			name := widget.NewLabel("")
			name.TextStyle.Bold = true
			name.Truncation = fyne.TextTruncateEllipsis
			state := widget.NewLabel("")
			info := container.NewBorder(nil, nil, nil, state, name)
			return card(container.NewBorder(nil, newLineBar(), widget.NewIcon(theme.MediaVideoIcon()), nil, info))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			it := items[i]
			row := o.(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*fyne.Container)
			info := row.Objects[0].(*fyne.Container)
			bar := row.Objects[1].(*fyne.Container)
			name := info.Objects[0].(*widget.Label)
			state := info.Objects[1].(*widget.Label)

			name.SetText(filepath.Base(it.path))
			var barColor color.Color = colAccent
			state.Importance = widget.LowImportance
			text := it.state
			switch it.state {
			case "변환 중":
				text, state.Importance = fmt.Sprintf("변환 중  %d%%", int(it.progress*100)), widget.HighImportance
			case "완료":
				state.Importance, barColor = widget.SuccessImportance, colSuccess
			case "실패":
				state.Importance, barColor = widget.DangerImportance, colError
			}
			state.SetText(text)
			info.Refresh() // 상태 라벨 폭이 바뀌므로 행 레이아웃 재계산
			setLineBar(bar, it.progress, barColor)
		})
	list.OnSelected = func(i widget.ListItemID) { // 실패 항목을 누르면 오류 내용 표시
		list.Unselect(i)
		if items[i].detail != "" {
			d := widget.NewLabel(items[i].detail)
			d.Wrapping = fyne.TextWrapWord
			dialog.ShowCustom(filepath.Base(items[i].path), "닫기", container.NewGridWrap(fyne.NewSize(480, 200), container.NewVScroll(d)), w)
		}
	}

	body := container.NewStack(dropZone, list)
	summary := canvas.NewText("", colMuted)
	summary.TextSize = 12

	var startBtn, clearBtn *widget.Button
	refresh := func() {
		done := 0
		for _, it := range items {
			if it.state == "완료" {
				done++
			}
		}
		if len(items) == 0 {
			dropZone.Show()
			list.Hide()
			summary.Text = "대기 중인 파일 없음"
		} else {
			dropZone.Hide()
			list.Show()
			summary.Text = fmt.Sprintf("%d개 파일 · %d개 완료", len(items), done)
		}
		summary.Refresh()
		list.Refresh()
	}

	clearBtn = widget.NewButtonWithIcon("목록 비우기", theme.DeleteIcon(), func() {
		items = nil
		refresh()
	})
	clearBtn.Importance = widget.LowImportance
	startBtn = widget.NewButtonWithIcon("변환 시작", theme.MediaPlayIcon(), func() {
		if running || len(items) == 0 {
			return
		}
		running = true
		startBtn.Disable()
		clearBtn.Disable()
		go func() {
			for i, it := range items {
				if it.state == "완료" {
					continue
				}
				fyne.Do(func() { it.state, it.progress, it.detail = "변환 중", 0, ""; list.RefreshItem(i) })
				out, err := convert(ctx, it.path, func(p float64) {
					fyne.Do(func() { it.progress = p; list.RefreshItem(i) })
				})
				fyne.Do(func() {
					if err != nil {
						it.state, it.detail = "실패", err.Error()
					} else {
						it.state, it.progress, it.detail = "완료", 1, "출력 폴더: "+out
					}
					refresh()
				})
			}
			fyne.Do(func() {
				running = false
				startBtn.Enable()
				clearBtn.Enable()
			})
		}()
	})
	startBtn.Importance = widget.HighImportance

	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		if running {
			return
		}
		for _, u := range uris {
			if strings.EqualFold(filepath.Ext(u.Path()), ".mp4") {
				items = append(items, &item{path: filepath.FromSlash(u.Path()), state: "대기"})
			}
		}
		refresh()
	})

	footer := container.NewHBox(container.NewCenter(summary), layout.NewSpacer(), clearBtn, startBtn)
	w.SetContent(container.NewPadded(container.NewBorder(
		container.NewVBox(header, layout.NewSpacer()), footer, nil, nil, body)))
	refresh()
}

// 얇은 진행바: track 위에 frac 비율만큼 fill을 그린다.
type lineBar struct{ frac float64 }

func (*lineBar) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(0, 4) }

func (l *lineBar) Layout(o []fyne.CanvasObject, s fyne.Size) {
	o[0].Resize(s)
	o[1].Resize(fyne.NewSize(s.Width*float32(l.frac), s.Height))
}

func newLineBar() *fyne.Container {
	track := canvas.NewRectangle(colBorder)
	track.CornerRadius = 2
	fill := canvas.NewRectangle(colAccent)
	fill.CornerRadius = 2
	return container.New(&lineBar{}, track, fill)
}

func setLineBar(c *fyne.Container, frac float64, col color.Color) {
	c.Layout.(*lineBar).frac = frac
	c.Objects[1].(*canvas.Rectangle).FillColor = col
	c.Refresh()
}
