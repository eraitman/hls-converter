package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("HLS 변환기")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 창을 닫으면 진행 중인 ffmpeg도 종료

	var files, status []string
	running := false

	list := widget.NewList(
		func() int { return len(files) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fmt.Sprintf("[%s] %s", status[i], filepath.Base(files[i])))
		})
	info := widget.NewLabel("mp4 파일을 이 창에 끌어다 놓으세요.")

	var startBtn, clearBtn *widget.Button
	clearBtn = widget.NewButton("목록 비우기", func() {
		files, status = nil, nil
		list.Refresh()
	})
	startBtn = widget.NewButton("변환 시작", func() {
		if running || len(files) == 0 {
			return
		}
		running = true
		startBtn.Disable()
		clearBtn.Disable()
		go func() {
			for i, f := range files {
				if status[i] == "완료" {
					continue
				}
				out, err := convert(ctx, f, func(p float64) {
					fyne.Do(func() { status[i] = fmt.Sprintf("%d%%", int(p*100)); list.RefreshItem(i) })
				})
				fyne.Do(func() {
					if err != nil {
						status[i] = "실패"
						info.SetText(filepath.Base(f) + ": " + err.Error())
					} else {
						status[i] = "완료"
						info.SetText("출력: " + out)
					}
					list.RefreshItem(i)
				})
			}
			fyne.Do(func() {
				running = false
				startBtn.Enable()
				clearBtn.Enable()
			})
		}()
	})

	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		if running {
			return
		}
		for _, u := range uris {
			if strings.EqualFold(filepath.Ext(u.Path()), ".mp4") {
				files = append(files, filepath.FromSlash(u.Path()))
				status = append(status, "대기")
			}
		}
		list.Refresh()
	})

	info.Wrapping = fyne.TextWrapWord
	w.SetContent(container.NewBorder(nil, container.NewVBox(info, container.NewHBox(startBtn, clearBtn)), nil, nil, list))
	w.Resize(fyne.NewSize(560, 420))
	w.ShowAndRun()
}
