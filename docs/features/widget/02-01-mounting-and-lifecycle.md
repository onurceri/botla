> Extracted from: widget.md — Section: 2.1 Mounting & Lifecycle

## 2.1 Mounting & Lifecycle

- [ ] `mount()` creates a Shadow DOM root on the target element
- [ ] `mount()` injects the widget stylesheet into the shadow root
- [ ] `mount()` reads all `data-*` attributes from the target element
- [ ] `mount()` falls back to `document.body` if no target element is specified
- [ ] `mount()` appends a container `<div>` to the shadow root
- [ ] `unmount()` removes the shadow root and all event listeners
- [ ] `unmount()` cleans up all timers (pulse animation, polling)
- [ ] Multiple widget instances on the same page are supported
- [ ] Widget errors are caught and logged without breaking the host page
