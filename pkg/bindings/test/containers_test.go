package test_bindings

import (
	"net/http"
	"strconv"
	"time"

	"github.com/containers/libpod/pkg/bindings"
	"github.com/containers/libpod/pkg/bindings/containers"
	"github.com/containers/libpod/test/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Podman containers ", func() {
	var (
		bt        *bindingTest
		s         *gexec.Session
		err       error
		falseFlag bool = false
		trueFlag  bool = true
	)

	BeforeEach(func() {
		bt = newBindingTest()
		bt.RestoreImagesFromCache()
		s = bt.startAPIService()
		time.Sleep(1 * time.Second)
		err := bt.NewConnection()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		s.Kill()
		//bt.cleanup()
	})

	It("podman pause a bogus container", func() {
		// Pausing bogus container should return 404
		err = containers.Pause(bt.conn, "foobar")
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusNotFound))
	})

	It("podman unpause a bogus container", func() {
		// Unpausing bogus container should return 404
		err = containers.Unpause(bt.conn, "foobar")
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusNotFound))
	})

	It("podman pause a running container by name", func() {
		// Pausing by name should work
		var name = "top"
		_, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, name)
		Expect(err).To(BeNil())

		// Ensure container is paused
		data, err := containers.Inspect(bt.conn, name, nil)
		Expect(err).To(BeNil())
		Expect(data.State.Status).To(Equal("paused"))
	})

	It("podman pause a running container by id", func() {
		// Pausing by id should work
		var name = "top"
		cid, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, cid)
		Expect(err).To(BeNil())

		// Ensure container is paused
		data, err := containers.Inspect(bt.conn, cid, nil)
		Expect(err).To(BeNil())
		Expect(data.State.Status).To(Equal("paused"))
	})

	It("podman unpause a running container by name", func() {
		// Unpausing by name should work
		var name = "top"
		_, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, name)
		Expect(err).To(BeNil())
		err = containers.Unpause(bt.conn, name)
		Expect(err).To(BeNil())

		// Ensure container is unpaused
		data, err := containers.Inspect(bt.conn, name, nil)
		Expect(err).To(BeNil())
		Expect(data.State.Status).To(Equal("running"))
	})

	It("podman unpause a running container by ID", func() {
		// Unpausing by ID should work
		var name = "top"
		_, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		// Pause by name
		err = containers.Pause(bt.conn, name)
		//paused := "paused"
		//_, err = containers.Wait(bt.conn, cid, &paused)
		//Expect(err).To(BeNil())
		err = containers.Unpause(bt.conn, name)
		Expect(err).To(BeNil())

		// Ensure container is unpaused
		data, err := containers.Inspect(bt.conn, name, nil)
		Expect(err).To(BeNil())
		Expect(data.State.Status).To(Equal("running"))
	})

	It("podman pause a paused container by name", func() {
		// Pausing a paused container by name should fail
		var name = "top"
		_, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, name)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, name)
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusInternalServerError))
	})

	It("podman pause a paused container by id", func() {
		// Pausing a paused container by id should fail
		var name = "top"
		cid, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, cid)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, cid)
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusInternalServerError))
	})

	It("podman pause a stopped container by name", func() {
		// Pausing a stopped container by name should fail
		var name = "top"
		_, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Stop(bt.conn, name, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, name)
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusInternalServerError))
	})

	It("podman pause a stopped container by id", func() {
		// Pausing a stopped container by id should fail
		var name = "top"
		cid, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Stop(bt.conn, cid, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, cid)
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusInternalServerError))
	})

	It("podman remove a paused container by id without force", func() {
		// Removing a paused container without force should fail
		var name = "top"
		cid, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, cid)
		Expect(err).To(BeNil())
		err = containers.Remove(bt.conn, cid, &falseFlag, &falseFlag)
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusInternalServerError))
	})

	It("podman remove a paused container by id with force", func() {
		// FIXME: Skip on F31 and later
		host := utils.GetHostDistributionInfo()
		osVer, err := strconv.Atoi(host.Version)
		Expect(err).To(BeNil())
		if host.Distribution == "fedora" && osVer >= 31 {
			Skip("FIXME: https://github.com/containers/libpod/issues/5325")
		}

		// Removing a paused container with force should work
		var name = "top"
		cid, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, cid)
		Expect(err).To(BeNil())
		err = containers.Remove(bt.conn, cid, &trueFlag, &falseFlag)
		Expect(err).To(BeNil())
	})

	It("podman stop a paused container by name", func() {
		// Stopping a paused container by name should fail
		var name = "top"
		_, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, name)
		Expect(err).To(BeNil())
		err = containers.Stop(bt.conn, name, nil)
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusInternalServerError))
	})

	It("podman stop a paused container by id", func() {
		// Stopping a paused container by id should fail
		var name = "top"
		cid, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Pause(bt.conn, cid)
		Expect(err).To(BeNil())
		err = containers.Stop(bt.conn, cid, nil)
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusInternalServerError))
	})

	It("podman stop a running container by name", func() {
		// Stopping a running container by name should work
		var name = "top"
		_, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Stop(bt.conn, name, nil)
		Expect(err).To(BeNil())

		// Ensure container is stopped
		data, err := containers.Inspect(bt.conn, name, nil)
		Expect(err).To(BeNil())
		Expect(isStopped(data.State.Status)).To(BeTrue())
	})

	It("podman stop a running container by ID", func() {
		// Stopping a running container by ID should work
		var name = "top"
		cid, err := bt.RunTopContainer(&name, &falseFlag, nil)
		Expect(err).To(BeNil())
		err = containers.Stop(bt.conn, cid, nil)
		Expect(err).To(BeNil())

		// Ensure container is stopped
		data, err := containers.Inspect(bt.conn, name, nil)
		Expect(err).To(BeNil())
		Expect(isStopped(data.State.Status)).To(BeTrue())
	})

	It("podman wait no condition", func() {
		var (
			name           = "top"
			exitCode int32 = -1
		)
		_, err := containers.Wait(bt.conn, "foobar", nil)
		Expect(err).ToNot(BeNil())
		code, _ := bindings.CheckResponseCode(err)
		Expect(code).To(BeNumerically("==", http.StatusNotFound))

		errChan := make(chan error)
		_, err = bt.RunTopContainer(&name, nil, nil)
		Expect(err).To(BeNil())
		go func() {
			exitCode, err = containers.Wait(bt.conn, name, nil)
			errChan <- err
			close(errChan)
		}()
		err = containers.Stop(bt.conn, name, nil)
		Expect(err).To(BeNil())
		wait := <-errChan
		Expect(wait).To(BeNil())
		Expect(exitCode).To(BeNumerically("==", 143))
	})

	It("podman wait to pause|unpause condition", func() {
		var (
			name           = "top"
			exitCode int32 = -1
			pause          = "paused"
			unpause        = "running"
		)
		errChan := make(chan error)
		_, err := bt.RunTopContainer(&name, nil, nil)
		Expect(err).To(BeNil())
		go func() {
			exitCode, err = containers.Wait(bt.conn, name, &pause)
			errChan <- err
			close(errChan)
		}()
		err = containers.Pause(bt.conn, name)
		Expect(err).To(BeNil())
		wait := <-errChan
		Expect(wait).To(BeNil())
		Expect(exitCode).To(BeNumerically("==", -1))

		errChan = make(chan error)
		go func() {
			exitCode, err = containers.Wait(bt.conn, name, &unpause)
			errChan <- err
			close(errChan)
		}()
		err = containers.Unpause(bt.conn, name)
		Expect(err).To(BeNil())
		unPausewait := <-errChan
		Expect(unPausewait).To(BeNil())
		Expect(exitCode).To(BeNumerically("==", -1))
	})

})
