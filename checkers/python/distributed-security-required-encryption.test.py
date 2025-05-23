# from https://github.com/apache/airflow/commit/80a3d6ac78c5c13abb8826b9dcbe0529f60fed81/

# -*- coding: utf-8 -*-
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

import distributed
import subprocess
import warnings

from airflow import configuration
from airflow.executors.base_executor import BaseExecutor


class DaskExecutor(BaseExecutor):
    """
    DaskExecutor submits tasks to a Dask Distributed cluster.
    """
    def __init__(self, cluster_address=None):
        if cluster_address is None:
            cluster_address = configuration.conf.get('dask', 'cluster_address')
        if not cluster_address:
            raise ValueError(
                'Please provide a Dask cluster address in airflow.cfg')
        self.cluster_address = cluster_address
        # ssl / tls parameters
        self.tls_ca = configuration.get('dask', 'tls_ca')
        self.tls_key = configuration.get('dask', 'tls_key')
        self.tls_cert = configuration.get('dask', 'tls_cert')
        super(DaskExecutor, self).__init__(parallelism=0)

    def start(self):
        if self.tls_ca or self.tls_key or self.tls_cert:
            from distributed.security import Security
            # <expect-error>
            security = Security(
                tls_client_key=self.tls_key,
                tls_client_cert=self.tls_cert,
                tls_ca_file=self.tls_ca,
                require_encryption=False,
            )
        else:
            security = None

        self.client = distributed.Client(self.cluster_address, security=security)
        self.futures = {}

    def execute_async(self, key, command, queue=None, executor_config=None):
        if queue is not None:
            warnings.warn(
                'DaskExecutor does not support queues. '
                'All tasks will be run in the same cluster'
            )

        def airflow_run():
            return subprocess.check_call(command, shell=True, close_fds=True)

        future = self.client.submit(airflow_run, pure=False)
        self.futures[future] = key

    def _process_future(self, future):
        if future.done():
            key = self.futures[future]
            if future.exception():
                self.log.error("Failed to execute task: %s", repr(future.exception()))
                self.fail(key)
            elif future.cancelled():
                self.log.error("Failed to execute task")
                self.fail(key)
            else:
                self.success(key)
            self.futures.pop(future)

    def sync(self):
        # make a copy so futures can be popped during iteration
        for future in self.futures.copy():
            self._process_future(future)

    def end(self):
        for future in distributed.as_completed(self.futures.copy()):
            self._process_future(future)

    def terminate(self):
        self.client.cancel(self.futures.keys())
        self.end()

    def create_security_context(self):
        # <expect-error>
        return distributed.security.Security(
            tls_ca_file="ca.pem",
            tls_cert="cert.pem",
            require_encryption=False,
        )
