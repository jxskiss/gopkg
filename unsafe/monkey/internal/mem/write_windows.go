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
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal/common"
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/internal/mem/prot"
)

func Write(addr uintptr, data []byte) error {
	if err := prot.MProtectRWX(addr); err != nil {
		return err
	}
	copy(common.BytesOf(addr, len(data)), data)
	return nil
}
