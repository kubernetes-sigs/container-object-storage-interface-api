#!/usr/bin/env bash

cat > /tmp/LICENSE_TEMPLATE << EOF
Copyright 2021 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
EOF

# install addlicense
GO111MODULE=off go get github.com/google/addlicense

# apply license to all go files
find . | grep .go$ | xargs addlicense -f /tmp/LICENSE_TEMPLATE
