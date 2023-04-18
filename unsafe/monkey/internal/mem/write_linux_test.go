/*
 * Copyright 2022 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mem

import (
	"fmt"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal/common"
)

func Test_write(t *testing.T) {
	var a uint16 = 0x0
	var b uint32 = 0xffffffff
	fmt.Printf("a=%x,b=%x\n", a, b)

	arr := (*[4]byte)(unsafe.Pointer(&a))
	arr[2] = 0xa5
	fmt.Printf("a=%x,b=%x,aSlice=%x\n", a, b, arr)

	target := uintptr(unsafe.Pointer(&a))
	data := uintptr(unsafe.Pointer(&b))
	res := write(target, data, 3, common.PageOf(target), common.PageSize(), syscall.PROT_READ|syscall.PROT_WRITE)
	fmt.Printf("a=%x,b=%x,aSlice=%x\n", a, b, arr)
	assert.Equal(t, 0, res)
	assert.Equal(t, 0xffff, a)
	assert.Equal(t, 0xffffffff, b)
	assert.Equal(t, [4]byte{0xff, 0xff, 0xff, 0x00}, *arr)
}
