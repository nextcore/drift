package overlay

import (
	"sync"

	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/graphics"
	"github.com/go-drift/drift/pkg/theme"
	"github.com/go-drift/drift/pkg/widgets"
)

// DialogBuilder creates dialog content given a [core.BuildContext] and a
// dismiss function. Call dismiss to close the dialog programmatically.
//
// The BuildContext passed to the builder comes from the overlay entry, so
// theme lookups ([theme.ThemeOf], [theme.UseTheme]) work as expected.
type DialogBuilder func(ctx core.BuildContext, dismiss func()) core.Widget

// DialogOptions configures [ShowDialog].
type DialogOptions struct {
	// Builder creates the dialog content widget. Required.
	//
	// The dismiss function passed to the builder removes both the barrier
	// and dialog entries from the overlay. It is safe to call multiple times.
	Builder DialogBuilder

	// Persistent prevents the barrier tap from dismissing the dialog.
	// When true, the user must interact with the dialog content (e.g.,
	// tap a button) to dismiss it. Default is false (barrier tap dismisses).
	Persistent bool

	// BarrierColor is the semi-transparent color drawn behind the dialog.
	// Zero value is fully transparent. Set explicitly for a visible scrim.
	BarrierColor graphics.Color
}

// ShowDialog displays a modal dialog above the nearest [Overlay].
//
// It creates two overlay entries: a [ModalBarrier] that absorbs touches behind
// the dialog, and a centered content entry built by opts.Builder. The dialog
// entry is marked Opaque so hit tests do not reach the page content below.
//
// The returned dismiss function removes both entries from the overlay. It is
// idempotent: calling it more than once is a safe no-op.
//
// ShowDialog must be called with a valid [core.BuildContext] that has an
// [Overlay] ancestor. Calling it from an async callback after the originating
// widget has been disposed may produce unexpected results; capture the dismiss
// function during build or from a gesture handler instead.
//
// If no [Overlay] ancestor exists in the widget tree or Builder is nil,
// ShowDialog returns a no-op dismiss function without inserting any entries.
//
// For a standard alert with string title/content and themed buttons, use
// [ShowAlertDialog] instead.
//
// Example with custom dialog content:
//
//	dismiss := overlay.ShowDialog(ctx, overlay.DialogOptions{
//	    BarrierColor: graphics.RGBA(0, 0, 0, 0.5),
//	    Builder: func(ctx core.BuildContext, dismiss func()) core.Widget {
//	        textTheme := theme.ThemeOf(ctx).TextTheme
//	        return overlay.Dialog{
//	            Child: widgets.Column{
//	                MainAxisSize: widgets.MainAxisSizeMin,
//	                Children: []core.Widget{
//	                    theme.TextOf(ctx, "Title", textTheme.HeadlineSmall),
//	                    widgets.VSpace(16),
//	                    theme.TextOf(ctx, "Body text", textTheme.BodyMedium),
//	                    widgets.VSpace(24),
//	                    theme.ButtonOf(ctx, "OK", dismiss),
//	                },
//	            },
//	        }
//	    },
//	})
//
// Example with persistent dialog (no barrier dismiss):
//
//	overlay.ShowDialog(ctx, overlay.DialogOptions{
//	    BarrierColor: graphics.RGBA(0, 0, 0, 0.5),
//	    Persistent:   true,
//	    Builder: func(ctx core.BuildContext, dismiss func()) core.Widget {
//	        textTheme := theme.ThemeOf(ctx).TextTheme
//	        return overlay.AlertDialog{
//	            Title:   theme.TextOf(ctx, "Processing", textTheme.HeadlineSmall),
//	            Content: widgets.CircularProgressIndicator{},
//	        }
//	    },
//	})
func ShowDialog(ctx core.BuildContext, opts DialogOptions) (dismiss func()) {
	ov := OverlayOf(ctx)
	if ov == nil || opts.Builder == nil {
		return func() {}
	}

	var once sync.Once
	var barrierEntry, dialogEntry *OverlayEntry

	// sync.Once guards against concurrent dismiss calls. OverlayEntry.Remove
	// is itself idempotent (no-op when overlay is nil), so external removal
	// (e.g., overlay dispose) does not cause issues.
	dismiss = func() {
		once.Do(func() {
			barrierEntry.Remove()
			dialogEntry.Remove()
		})
	}

	barrierEntry = NewOverlayEntry(func(ctx core.BuildContext) core.Widget {
		return ModalBarrier{
			Color:       opts.BarrierColor,
			Dismissible: !opts.Persistent,
			OnDismiss:   dismiss,
		}
	})

	dialogEntry = NewOverlayEntry(func(ctx core.BuildContext) core.Widget {
		return widgets.Center{
			Child: opts.Builder(ctx, dismiss),
		}
	})
	// Opaque blocks hits from reaching the page content (Overlay.Child) but
	// does not prevent other overlay entries from receiving hits. This lets
	// the barrier entry (below the dialog) still handle dismiss taps.
	dialogEntry.Opaque = true

	ov.InsertAll([]*OverlayEntry{barrierEntry, dialogEntry}, nil, nil)

	return dismiss
}

// AlertDialogOptions configures [ShowAlertDialog].
type AlertDialogOptions struct {
	// Title is the dialog heading text. Rendered with HeadlineSmall style.
	// Empty string omits the title.
	Title string

	// Content is the dialog body text. Rendered with BodyMedium style.
	// Empty string omits the content.
	Content string

	// ConfirmLabel is the label for the confirm (primary) action button.
	// Empty string omits the confirm button.
	ConfirmLabel string

	// OnConfirm is called when the confirm button is tapped, before the
	// dialog is dismissed. May be nil if no callback is needed.
	OnConfirm func()

	// CancelLabel is the label for the cancel (secondary) action button.
	// Empty string omits the cancel button.
	CancelLabel string

	// OnCancel is called when the cancel button is tapped, before the
	// dialog is dismissed. May be nil if no callback is needed.
	OnCancel func()

	// Destructive styles the confirm button with the theme's Error color
	// instead of Primary, signaling a dangerous action such as deletion.
	Destructive bool

	// Persistent prevents the barrier tap from dismissing the dialog.
	// Passed through to [DialogOptions.Persistent].
	Persistent bool
}

// ShowAlertDialog displays a standard alert dialog with themed title, content,
// and action buttons.
//
// This is a convenience wrapper around [ShowDialog] and [AlertDialog] that
// handles text styling and button construction. The cancel button uses
// SecondaryContainer colors and the confirm button uses Primary colors
// (or Error colors when Destructive is true).
//
// Both buttons dismiss the dialog after calling their respective callbacks.
// The returned dismiss function allows programmatic dismissal from outside
// the dialog.
//
// Example:
//
//	overlay.ShowAlertDialog(ctx, overlay.AlertDialogOptions{
//	    Title:        "Delete item?",
//	    Content:      "This action cannot be undone.",
//	    ConfirmLabel: "Delete",
//	    OnConfirm:    func() { deleteItem() },
//	    CancelLabel:  "Cancel",
//	    Destructive:  true,
//	})
//
// Example with confirm only:
//
//	overlay.ShowAlertDialog(ctx, overlay.AlertDialogOptions{
//	    Title:        "Update available",
//	    Content:      "A new version is ready to install.",
//	    ConfirmLabel: "OK",
//	})
func ShowAlertDialog(ctx core.BuildContext, opts AlertDialogOptions) (dismiss func()) {
	return ShowDialog(ctx, DialogOptions{
		Persistent:   opts.Persistent,
		BarrierColor: theme.ThemeOf(ctx).ColorScheme.Scrim.WithAlpha(0.5),
		Builder: func(ctx core.BuildContext, dismiss func()) core.Widget {
			th := theme.ThemeOf(ctx)
			colors := th.ColorScheme
			textTheme := th.TextTheme

			var titleWidget core.Widget
			if opts.Title != "" {
				titleWidget = theme.TextOf(ctx, opts.Title, textTheme.HeadlineSmall)
			}

			var contentWidget core.Widget
			if opts.Content != "" {
				contentWidget = theme.TextOf(ctx, opts.Content, textTheme.BodyMedium)
			}

			var actions []core.Widget

			if opts.CancelLabel != "" {
				cancelCallback := opts.OnCancel
				actions = append(actions, theme.ButtonOf(ctx, opts.CancelLabel, func() {
					if cancelCallback != nil {
						cancelCallback()
					}
					dismiss()
				}).WithColor(colors.SecondaryContainer, colors.OnSecondaryContainer))
			}

			if opts.ConfirmLabel != "" {
				confirmCallback := opts.OnConfirm
				bg := colors.Primary
				fg := colors.OnPrimary
				if opts.Destructive {
					bg = colors.Error
					fg = colors.OnError
				}
				actions = append(actions, theme.ButtonOf(ctx, opts.ConfirmLabel, func() {
					if confirmCallback != nil {
						confirmCallback()
					}
					dismiss()
				}).WithColor(bg, fg))
			}

			return AlertDialog{
				Title:   titleWidget,
				Content: contentWidget,
				Actions: actions,
			}
		},
	})
}
