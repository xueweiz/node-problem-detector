#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script is for installing node problem detector (NPD) on a running node
# in standalone mode, as a setup for NPD e2e tests.

set -o errexit
set -o nounset
set -o pipefail

function print-help() {
  echo "Usage: install.sh [flags] [command]"
  echo
  echo "Available flags:"
  echo "  -t [TARBALL]     Specify the path of the NPD tarball (generated by 'make build-tar')."
  echo
  echo "Available commands:"
  echo "  help     Print this help message"
  echo "  install  Installs NPD to the this machine"
  echo
  echo "Examples:"
  echo "  install.sh help"
  echo "  install.sh -t /tmp/npd.tar.gz install"
}

function install-npd() {
  if [[ -z "${TARBALL}" ]]; then
  	echo "ERROR: tarball flag is missing."
  	exit 1
  fi

  workdir=$(mktemp -d)
  tar -xf "${TARBALL}" --directory "${workdir}"
  ls "${workdir}"
  exit 0
  
  installdir=/etc/node_problem_detector/
  mkdir -p "${installdir}"

  cp "${workdir}/bin/node-problem-detector" "${installdir}"
  cp "${workdir}/config/*" "${installdir}"

  rm -rf "${workdir}"

  cat <<EOF >/etc/systemd/system/node-problem-detector.service
[Unit]
Description=Node problem detector
Wants=local-fs.target
After=local-fs.target

[Service]
Restart=always
RestartSec=10
ExecStart=/etc/node-problem-detector --config.system-stats-monitor=/etc/node_problem_detector/system-stats-monitor.json --config.system-log-monitor=/etc/node_problem_detector/kernel-monitor.json --enable-k8s-exporter=false --prometheus-port=20257 --alsologtostderr

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl stop node-problem-detector.service || true
  systemctl start node-problem-detector.service
}

function main() {
  case ${1:-} in
  help) print-help;;
  install) install-npd;;
  *) print-help;;
  esac
}

TARBALL=""

while getopts "t:" opt; do
  case ${opt} in
    t) TARBALL="${OPTARG}";;
  esac
done
shift "$((OPTIND-1))"


main "${@}"