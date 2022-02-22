#!/bin/bash
#
# Copyright (c) 2021 Red Hat
#
# SPDX-License-Identifier: Apache-2.0
#

# Webhook namespace.
WEBHOOK_NS=${WEBHOOK_NS:-"default"}
# Webhook Pod name.
WEBHOOK_NAME=${WEBHOOK_NAME:-"pod-mutate"}
# Webhook service name.
WEBHOOK_SVC="${WEBHOOK_NAME}-webhook"