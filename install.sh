#!/bin/bash

_OS_NAME=$(uname -s)
_CPU_ARCH=$(uname -m)

OS_NAME="unknown"
CPU_ARCH="unknown"

case "${_OS_NAME}" in
  Linux*)   OS_NAME="linux";;
  Darwin*)  OS_NAME="darwin";;
  CYGWIN*)  OS_NAME="windows";;
  MINGW*)   OS_NAME="windows";;
esac

case "${_CPU_ARCH}" in
  x86_64)  CPU_ARCH="amd64";;
  arm64)   CPU_ARCH="arm64";;
esac

if [ "${OS_NAME}" == "unknown" ]; then
  echo "Unsupported OS: ${_OS_NAME}"
  exit 1
fi

if [ "${CPU_ARCH}" == "unknown" ]; then
  echo "Unsupported CPU architecture: ${_CPU_ARCH}"
  exit 1
fi

echo "Downloading az-mpf for ${OS_NAME} ${CPU_ARCH}..."

_LATEST_RELEASE=$(curl -s "https://api.github.com/repos/manisbindra/az-mpf/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")')

echo "Latest release: ${_LATEST_RELEASE}"

curl -L https://github.com/manisbindra/az-mpf/releases/download/${_LATEST_RELEASE}/az-mpf-${OS_NAME}-${CPU_ARCH}.exe -o /usr/local/bin/az-mpf

chmod +x /usr/local/bin/az-mpf

cat << EOF >> ~/.bashrc
function az() {
  if [[ "\${1}" == "mpf" ]]; then
    declare -A params
    local key

    shift 1

    while [ \$# -gt 0 ]; do
      key="\$1"
      case "\$key" in
        --subscription-id)
          params[subscription-id]="\$2"
          shift 2
          ;;
        --service-principal-client-id)
          params[service-principal-client-id]="\$2"
          shift 2
          ;;
        --service-principal-object-id)
          params[service-principal-object-id]="\$2"
          shift 2
          ;;
        --service-principal-client-secret)
          params[service-principal-client-secret]="\$2"
          shift 2
          ;;
        --tenant-id)
          params[tenant-id]="\$2"
          shift 2
          ;;
        --template-file)
          params[template-file]="\$2"
          shift 2
          ;;
        --parameters-file)
          params[parameters-file]="\$2"
          shift 2
          ;;
        --show-detailed-output)
          params[show-detailed-output]="\$2"
          shift 2
          ;;
        --json-output)
          params[json-output]="\$2"
          shift 2
          ;;
        *)
          echo "Invalid option: \${key}" 1>&2
          exit 1
          ;;
      esac
    done

    SUBSCRIPTION_ID=\${params[subscription-id]}
    SERVICE_PRINCIPAL_CLIENT_ID=\${params[service-principal-client-id]}
    SERVICE_PRINCIPAL_OBJECT_ID=\${params[service-principal-object-id]}
    SERVICE_PRINCIPAL_CLIENT_SECRET=\${params[service-principal-client-secret]}
    TENANT_ID=\${params[tenant-id]}
    TEMPLATE_FILE=\${params[template-file]}
    PARAMETERS_FILE=\${params[parameters-file]}
    SHOW_DETAILED_OUTPUT=\${params[show-detailed-output]}
    JSON_OUTPUT=\${params[json-output]}

    if [ -z "\${SUBSCRIPTION_ID}" ]; then
      echo "Missing required parameter: --subscription-id" 1>&2
      exit 1
    fi

    if [ -z "\${SERVICE_PRINCIPAL_CLIENT_ID}" ]; then
      echo "Missing required parameter: --service-principal-client-id" 1>&2
      exit 1
    fi

    if [ -z "\${SERVICE_PRINCIPAL_OBJECT_ID}" ]; then
      echo "Missing required parameter: --service-principal-object-id" 1>&2
      exit 1
    fi

    if [ -z "\${SERVICE_PRINCIPAL_CLIENT_SECRET}" ]; then
      echo "Missing required parameter: --service-principal-client-secret" 1>&2
      exit 1
    fi

    if [ -z "\${TENANT_ID}" ]; then
      echo "Missing required parameter: --tenant-id" 1>&2
      exit 1
    fi

    if [ -z "\${TEMPLATE_FILE}" ]; then
      echo "Missing required parameter: --template-file" 1>&2
      exit 1
    fi

    if [ -z "\${PARAMETERS_FILE}" ]; then
      echo "Missing required parameter: --parameters-file" 1>&2
      exit 1
    fi

    if [ ! -f "\${TEMPLATE_FILE}" ]; then
      echo "Template file not found: \${TEMPLATE_FILE}" 1>&2
      exit 1
    fi

    if [ ! -f "\${PARAMETERS_FILE}" ]; then
      echo "Parameters file not found: \${PARAMETERS_FILE}" 1>&2
      exit 1
    fi

    if [ -z "\${SHOW_DETAILED_OUTPUT}" ]; then
      SHOW_DETAILED_OUTPUT="false"
    fi

    if [ -z "\${JSON_OUTPUT}" ]; then
      JSON_OUTPUT="false"
    fi

    /usr/local/bin/az-mpf -subscriptionID "\${SUBSCRIPTION_ID}" -servicePrincipalClientID "\${SERVICE_PRINCIPAL_CLIENT_ID}" -servicePrincipalObjectID "\${SERVICE_PRINCIPAL_OBJECT_ID}" -servicePrincipalClientSecret "\${SERVICE_PRINCIPAL_CLIENT_SECRET}" -tenantID "\${TENANT_ID}" -templateFile "\${TEMPLATE_FILE}" -parametersFile "\${PARAMETERS_FILE}" -showDetailedOutput "\${SHOW_DETAILED_OUTPUT}" -jsonOutput "\${JSON_OUTPUT}"

    return 0;
  fi

  command az "\${@}"
}
EOF

exit 0;
