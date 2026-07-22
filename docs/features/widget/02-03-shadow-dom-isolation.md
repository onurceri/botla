> Extracted from: widget.md — Section: 2.3 Shadow DOM Isolation

## 2.3 Shadow DOM Isolation

- [ ] Styles defined in shadow root do not leak to the host page
- [ ] Host page styles do not affect widget rendering
- [ ] CSS reset is applied aggressively inside shadow root
- [ ] Font family is explicitly defined (does not inherit from host)
- [ ] `all: initial` is applied to the shadow root container
- [ ] z-index stacking is isolated from host page
- [ ] Focus is trapped inside open modals (confirm modal)
- [ ] Tab order within widget is correct
