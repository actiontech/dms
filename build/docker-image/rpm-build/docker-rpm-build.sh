#!/bin/bash
set -e "${VERBOSE:+-x}"

SPEC="/spec/$1"
if [[ -z ${SPEC} || ! -e ${SPEC} ]]; then
  echo "Usage: docker run [--rm]" \
    "--volume=/path/to/source:/src --volume=/path/to/spec-dir:/spec --workdir=/src" \
    "rpmbuild-image SPEC_FILE_NAME" >&2
  exit 2
fi

# pre-builddep hook for adding extra repos
if [[ -n ${PRE_BUILDDEP} ]]; then
  bash "${VERBOSE:+-x}" -c "${PRE_BUILDDEP}"
fi

TOPDIR="${HOME}/rpmbuild"

cp "${VERBOSE:+-v}" -a --reflink=auto /src/* "${TOPDIR}/SOURCES/"
cp "${VERBOSE:+-v}" -a --reflink=auto "${SPEC}" "${TOPDIR}/SPECS/"

SPEC="${TOPDIR}/SPECS/${1}"
DEFAULT_RPM_ARGS="-ba"

if [[ -n ${RPM_ARGS} ]]; then
  DEFAULT_RPM_ARGS=${RPM_ARGS}
fi

CMD="rpmbuild ${VERBOSE:+-v} ${DEFAULT_RPM_ARGS} ${SPEC}"
bash -c "${CMD}"

OUTDIR="/out"
cp "${VERBOSE:+-v}" -a --reflink=auto \
  ${TOPDIR}/RPMS "${OUTDIR}/"
