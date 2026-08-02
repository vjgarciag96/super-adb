# 13 — Clear stored serial when device disconnects

**What to build:** The session store (`~/.sadb/device`) should persist the selected device forever — but only as long as it is connected. When `Resolve` detects that the stored serial is no longer in the live device list, it should clear the store before falling through to normal resolution (auto-select or picker). This keeps the file from accumulating stale state across device swaps.

**Current behaviour:** The stale-serial guard (added in #10) correctly falls through to the picker when the stored device is missing, but leaves the old serial in the file. It gets overwritten only on the next successful pick.

**Desired behaviour:**
- Device connected → stored serial used, no prompt.
- Device disconnected → store cleared immediately, normal resolution runs (auto-select if one device, picker if multiple).
- Store is never left holding a serial that isn't currently connected.

**Blocked by:** 10 — Guard against stale stored serial (done)

**Status:** ready-for-agent

- [ ] When the stored serial is absent from the live device list, `cfg.Store.Save("")` is called before falling through
- [ ] `FileStore.Load` treats an empty file as "nothing stored" (already true — `TrimSpace("")` returns `""`)
- [ ] Test covers the stale-serial-clears-store path
- [ ] Existing stale-serial tests still pass
