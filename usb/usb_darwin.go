// Package usb provides USB bulk transfer access via macOS IOKit.
package usb

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <IOKit/IOKitLib.h>
#include <IOKit/usb/IOUSBLib.h>
#include <IOKit/IOCFPlugIn.h>
#include <CoreFoundation/CoreFoundation.h>

typedef struct {
	IOUSBInterfaceInterface **iface;
	UInt8                     pipeRef;
} usb_handle;

// find_device locates a USB device service by vendor and product ID.
static io_service_t find_device(UInt16 vid, UInt16 pid) {
	// Try both modern and legacy class names.
	const char *classes[] = {"IOUSBHostDevice", "IOUSBDevice", NULL};
	for (int i = 0; classes[i]; i++) {
		CFMutableDictionaryRef dict = IOServiceMatching(classes[i]);
		if (!dict) continue;

		SInt32 v = vid, p = pid;
		CFNumberRef vRef = CFNumberCreate(NULL, kCFNumberSInt32Type, &v);
		CFNumberRef pRef = CFNumberCreate(NULL, kCFNumberSInt32Type, &p);
		CFDictionarySetValue(dict, CFSTR("idVendor"), vRef);
		CFDictionarySetValue(dict, CFSTR("idProduct"), pRef);
		CFRelease(vRef);
		CFRelease(pRef);

		io_iterator_t iter;
		// IOServiceGetMatchingServices consumes dict.
		kern_return_t kr = IOServiceGetMatchingServices(0, dict, &iter);
		if (kr != KERN_SUCCESS) continue;

		io_service_t svc = IOIteratorNext(iter);
		IOObjectRelease(iter);
		if (svc) return svc;
	}
	return 0;
}

// try_open_interface attempts to open an interface service and check for
// a bulk OUT endpoint at the given address.
static IOReturn try_open_interface(io_service_t svc, UInt8 endpoint,
		IOUSBInterfaceInterface ***out, UInt8 *outPipe) {
	IOCFPlugInInterface **plug = NULL;
	SInt32 score;
	IOReturn kr = IOCreatePlugInInterfaceForService(svc,
		kIOUSBInterfaceUserClientTypeID, kIOCFPlugInInterfaceID, &plug, &score);
	if (kr != kIOReturnSuccess || !plug) return kr ? kr : kIOReturnError;

	IOUSBInterfaceInterface **iface = NULL;
	HRESULT hr = (*plug)->QueryInterface(plug,
		CFUUIDGetUUIDBytes(kIOUSBInterfaceInterfaceID), (LPVOID *)&iface);
	(*plug)->Release(plug);
	if (hr != S_OK || !iface) return kIOReturnError;

	// Skip Audio class interfaces (class 1) — those are MIDI, not display.
	UInt8 ifaceClass;
	kr = (*iface)->GetInterfaceClass(iface, &ifaceClass);
	if (kr == kIOReturnSuccess && ifaceClass == 1) {
		(*iface)->Release(iface);
		return kIOReturnNotFound;
	}

	kr = (*iface)->USBInterfaceOpen(iface);
	if (kr != kIOReturnSuccess) {
		(*iface)->Release(iface);
		return kr;
	}

	UInt8 numEP;
	kr = (*iface)->GetNumEndpoints(iface, &numEP);
	if (kr != kIOReturnSuccess) {
		(*iface)->USBInterfaceClose(iface);
		(*iface)->Release(iface);
		return kr;
	}

	for (UInt8 pipe = 1; pipe <= numEP; pipe++) {
		UInt8 dir, num, xferType, interval;
		UInt16 maxPkt;
		kr = (*iface)->GetPipeProperties(iface, pipe,
			&dir, &num, &xferType, &maxPkt, &interval);
		if (kr != kIOReturnSuccess) continue;
		if (dir == kUSBOut && xferType == kUSBBulk && num == endpoint) {
			*out = iface;
			*outPipe = pipe;
			return kIOReturnSuccess;
		}
	}

	(*iface)->USBInterfaceClose(iface);
	(*iface)->Release(iface);
	return kIOReturnNotFound;
}

// usb_open finds a USB device by VID/PID, then walks its child interface
// services to find one with a bulk OUT endpoint at the given address.
// Does NOT open the device itself (the kernel composite driver holds it).
static IOReturn usb_open(UInt16 vid, UInt16 pid, UInt8 endpoint, usb_handle *h) {
	io_service_t devSvc = find_device(vid, pid);
	if (!devSvc) return kIOReturnNotFound;

	// Walk child services (USB interfaces) in the IOService plane.
	io_iterator_t childIter;
	IOReturn kr = IORegistryEntryGetChildIterator(devSvc, kIOServicePlane, &childIter);
	IOObjectRelease(devSvc);
	if (kr != kIOReturnSuccess) return kr;

	io_service_t child;
	while ((child = IOIteratorNext(childIter))) {
		IOUSBInterfaceInterface **iface = NULL;
		UInt8 pipeRef = 0;
		kr = try_open_interface(child, endpoint, &iface, &pipeRef);
		IOObjectRelease(child);
		if (kr == kIOReturnSuccess) {
			IOObjectRelease(childIter);
			h->iface   = iface;
			h->pipeRef = pipeRef;
			return kIOReturnSuccess;
		}
	}
	IOObjectRelease(childIter);
	return kIOReturnNotFound;
}

static IOReturn usb_write(usb_handle *h, void *data, UInt32 len) {
	return (*h->iface)->WritePipe(h->iface, h->pipeRef, data, len);
}

static IOReturn usb_clear_stall(usb_handle *h) {
	return (*h->iface)->ClearPipeStallBothEnds(h->iface, h->pipeRef);
}

static IOReturn usb_close(usb_handle *h) {
	if (!h->iface) return kIOReturnSuccess;
	IOReturn kr = (*h->iface)->USBInterfaceClose(h->iface);
	(*h->iface)->Release(h->iface);
	h->iface = NULL;
	return kr;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Device represents an open USB device with a claimed bulk OUT endpoint.
type Device struct {
	handle C.usb_handle
}

// Open finds a USB device by vendor/product ID, and claims the interface
// containing the specified bulk OUT endpoint.
func Open(vendorID, productID uint16, endpoint uint8) (*Device, error) {
	d := &Device{}
	kr := C.usb_open(C.UInt16(vendorID), C.UInt16(productID), C.UInt8(endpoint), &d.handle)
	if kr != 0 {
		return nil, fmt.Errorf("usb: open %04x:%04x endpoint %d: %w", vendorID, productID, endpoint, ioError(kr))
	}
	return d, nil
}

// ErrPipeStalled indicates the USB pipe has stalled and was cleared.
// The caller should skip the current operation and retry on the next cycle.
var ErrPipeStalled = fmt.Errorf("usb: pipe stalled")

// Write sends data to the bulk OUT endpoint.
// If the pipe stalls, it is automatically cleared and ErrPipeStalled is returned.
func (d *Device) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	kr := C.usb_write(&d.handle, unsafe.Pointer(&data[0]), C.UInt32(len(data)))
	if kr == 0 {
		return len(data), nil
	}
	// 0xe00002eb = kIOUSBPipeStalled
	if uint32(kr) == 0xe00002eb {
		C.usb_clear_stall(&d.handle)
		return 0, ErrPipeStalled
	}
	return 0, fmt.Errorf("usb: write: %w", ioError(kr))
}

// Close releases the interface.
func (d *Device) Close() error {
	kr := C.usb_close(&d.handle)
	if kr != 0 {
		return fmt.Errorf("usb: close: %w", ioError(kr))
	}
	return nil
}

// ioError converts an IOKit return code to a Go error.
type ioError C.IOReturn

func (e ioError) Error() string {
	return fmt.Sprintf("iokit error 0x%08x", uint32(e))
}
