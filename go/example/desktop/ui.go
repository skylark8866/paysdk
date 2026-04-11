package main

import (
	"image"
	"image/color"
	"io"
	"strings"

	"gioui.org/font/gofont"
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type UI struct {
	theme *material.Theme

	createBtn   widget.Clickable
	queryBtn    widget.Clickable
	refundBtn   widget.Clickable
	copyBtn     widget.Clickable
	stopPollBtn widget.Clickable

	amountEditor widget.Editor
	titleEditor  widget.Editor

	qrImage *widget.Image
}

func NewUI() *UI {
	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	theme.Palette.ContrastBg = color.NRGBA{R: 0, G: 120, B: 215, A: 255}

	return &UI{
		theme: theme,
		qrImage: &widget.Image{
			Fit: widget.Unscaled,
		},
	}
}

func (ui *UI) Layout(gtx layout.Context, state *AppState) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(ui.layoutHeader),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutInput(gtx, state)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutButtons(gtx, state)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutQRCode(gtx, state)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutStatus(gtx, state)
		}),
	)
}

func (ui *UI) layoutHeader(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		title := material.H4(ui.theme, "XGDN Pay 测试工具")
		title.Color = ui.theme.Palette.ContrastBg
		return title.Layout(gtx)
	})
}

func (ui *UI) layoutInput(gtx layout.Context, state *AppState) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
					layout.Flexed(0.45, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Label(ui.theme, ui.theme.TextSize, "金额")
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								ui.amountEditor.SingleLine = true
								return material.Editor(ui.theme, &ui.amountEditor, "0.01").Layout(gtx)
							}),
						)
					}),
					layout.Flexed(0.45, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Label(ui.theme, ui.theme.TextSize, "标题")
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								ui.titleEditor.SingleLine = true
								return material.Editor(ui.theme, &ui.titleEditor, "测试商品").Layout(gtx)
							}),
						)
					}),
				)
			}),
		)
	})
}

func (ui *UI) layoutButtons(gtx layout.Context, state *AppState) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEvenly}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(ui.theme, &ui.createBtn, "创建订单")
				btn.Background = ui.theme.Palette.ContrastBg
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(ui.theme, &ui.queryBtn, "查询状态")
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(ui.theme, &ui.refundBtn, "申请退款")
				btn.Background = color.NRGBA{R: 220, G: 53, B: 69, A: 255}
				return btn.Layout(gtx)
			}),
		)
	})
}

func (ui *UI) layoutQRCode(gtx layout.Context, state *AppState) layout.Dimensions {
	qrImage := state.GetQRCode()
	if qrImage == nil {
		return layout.Dimensions{}
	}

	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				ui.qrImage.Src = paint.NewImageOp(qrImage)
				gtx.Constraints.Max = image.Pt(256, 256)
				gtx.Constraints.Min = image.Pt(256, 256)
				return ui.qrImage.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				orderNo, _, _, _, _ := state.GetOrderInfo()
				info := material.Body1(ui.theme, "订单号: "+orderNo)
				return info.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				_, outOrderNo, _, _, _ := state.GetOrderInfo()
				info := material.Body1(ui.theme, "业务单号: "+outOrderNo)
				return info.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(ui.theme, &ui.copyBtn, "复制支付链接")
				btn.Background = color.NRGBA{R: 40, G: 167, B: 69, A: 255}
				return btn.Layout(gtx)
			}),
		)
	})
}

func (ui *UI) layoutStatus(gtx layout.Context, state *AppState) layout.Dimensions {
	status := state.GetStatus()
	isPolling := state.IsPollingNow()

	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				bg := color.NRGBA{R: 240, G: 240, B: 240, A: 255}
				return ui.fillBackground(gtx, bg, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.H6(ui.theme, status)
						label.Alignment = text.Middle
						return label.Layout(gtx)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !isPolling {
					return layout.Dimensions{}
				}
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(ui.theme, &ui.stopPollBtn, "停止轮询")
						btn.Background = color.NRGBA{R: 255, G: 193, B: 7, A: 255}
						return btn.Layout(gtx)
					}),
				)
			}),
		)
	})
}

func (ui *UI) fillBackground(gtx layout.Context, c color.NRGBA, w layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	path := clip.Rect{Max: dims.Size}.Op()
	paint.FillShape(gtx.Ops, c, path)
	call.Add(gtx.Ops)

	return dims
}

func (ui *UI) HandleEvents(gtx layout.Context, state *AppState) {
	if ui.createBtn.Clicked(gtx) {
		go state.CreateOrder()
	}

	if ui.queryBtn.Clicked(gtx) {
		go state.QueryOrder()
	}

	if ui.refundBtn.Clicked(gtx) {
		go state.Refund()
	}

	if ui.copyBtn.Clicked(gtx) {
		_, _, codeURL, _, _ := state.GetOrderInfo()
		if codeURL != "" {
			gtx.Execute(clipboard.WriteCmd{Data: io.NopCloser(strings.NewReader(codeURL))})
		}
	}

	if ui.stopPollBtn.Clicked(gtx) {
		state.StopPolling()
		state.SetStatus("已停止轮询")
	}
}
