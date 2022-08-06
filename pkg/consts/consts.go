/* Copyright 2022 The Kubernetes Authors.
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

package consts

const (
	AccountNamePrefix = "ba-"
	BucketInfoPrefix  = "bc-"

	BABucketFinalizer = "cosi.objectstorage.k8s.io/bucketaccess-bucket-protection"
	BAFinalizer       = "cosi.objectstorage.k8s.io/bucketaccess-protection"
	BCFinalizer       = "cosi.objectstorage.k8s.io/bucketclaim-protection"
	BucketFinalizer   = "cosi.objectstorage.k8s.io/bucket-protection"
	SecretFinalizer   = "cosi.objectstorage.k8s.io/secret-protection"

	S3Key                      = "s3"
	AzureKey                   = "azure"
	S3SecretAccessKeyID        = "accessKeyID"
	S3SecretAccessSecretKey    = "accessSecretKey"
	AzureSecretAccessToken     = "accessToken"
	AzureSecretExpiryTimeStamp = "expiryTs"
	DefaultTimeFormat          = "2006-01-02 15:04:05.999999999 -0700 MST"
)
