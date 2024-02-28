#!/bin/bash

# Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# You may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit

if [ -z "${GOMOCK_VERSION:-}" ]; then
    echo "GOMOCK_VERSION must be set"
    exit 1
fi

if [ -z "${SWAGGER_VERSION:-}" ]; then
    echo "SWAGGER_VERSION must be set"
    exit 1
fi

if [ -z "${GOWRAP_VERSION:-}" ]; then
    echo "GOWRAP_VERSION must be set"
    exit 1
fi

echo "Downloading gomock $GOMOCK_VERSION"
go install -v github.com/golang/mock/mockgen@${GOMOCK_VERSION}
echo "verify mockgen installation"
${GOBIN}/mockgen -version

go install -v github.com/go-swagger/go-swagger/cmd/swagger@${SWAGGER_VERSION}
echo "verify swagger installation"
${GOBIN}/swagger version

echo "Downloading gowrap $GOWRAP_VERSION"
go install -v github.com/hexdigest/gowrap/cmd/gowrap@${GOWRAP_VERSION}
echo "Verifying gowrap installation"
${GOBIN}/gowrap help
[ $? -eq 0 ] && echo "gowrap is installed in ${GOBIN}" || echo "gowrap was not found in ${GOBIN}"
