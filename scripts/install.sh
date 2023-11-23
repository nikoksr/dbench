#!/bin/bash -
#===============================================================================
#
#   DESCRIPTION: dbench Installer Script. All credits to https://github.com/schollz/croc
#
#                This script installs dbench into a specified prefix.
#                Default prefix = /usr/local/bin
#
#       OPTIONS: -p, --prefix "${INSTALL_PREFIX}"
#                      Prefix to install dbench into.  Defaults to /usr/local/bin
#  REQUIREMENTS: bash, uname, tar/unzip, curl/wget, sudo (if not run
#                as root), install, mktemp, sha256sum/shasum/sha256
#
#          BUGS: ...hopefully not.  Please report.
#
#          Issues: https://github.com/nikoksr/dbench/issues
#===============================================================================
set -o nounset # Treat unset variables as an error

#-------------------------------------------------------------------------------
# DEFAULTS
#-------------------------------------------------------------------------------
PREFIX="${PREFIX:-}"
ANDROID_ROOT="${ANDROID_ROOT:-}"

# Termux on Android has ${PREFIX} set which already ends with '/usr'
if [[ -n "${ANDROID_ROOT}" && -n "${PREFIX}" ]]; then
	INSTALL_PREFIX="${PREFIX}/bin"
else
	INSTALL_PREFIX="/usr/local/bin"
fi

#-------------------------------------------------------------------------------
# FUNCTIONS
#-------------------------------------------------------------------------------

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  print_banner
#   DESCRIPTION:  Prints a banner
#    PARAMETERS:  none
#       RETURNS:  0
#-------------------------------------------------------------------------------
print_banner() {
	cat <<-'EOF'
		=================================================

		     ______  ______                    _
		    (______)(____  \                  | |
		     _     _ ____)  )_____ ____   ____| |__
		    | |   | |  __  (| ___ |  _ \ / ___)  _ \
		    | |__/ /| |__)  ) ____| | | ( (___| | | |
		    |_____/ |______/|_____)_| |_|\____)_| |_|

		     _                        _ _
		    | |             _        | | |
		    | |____   ___ _| |_ _____| | | _____  ____
		    | |  _ \ /___|_   _|____ | | || ___ |/ ___)
		    | | | | |___ | | |_/ ___ | | || ____| |
		    |_|_| |_(___/   \__)_____|\_)_)_____)_|

		==================================================
	EOF
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  print_help
#   DESCRIPTION:  Prints out a help message
#    PARAMETERS:  none
#       RETURNS:  0
#-------------------------------------------------------------------------------
print_help() {
	local help_header
	local help_message

	help_header="dbench Installer Script"
	help_message="Usage:
  -p INSTALL_PREFIX
      Prefix to install dbench into.  Directory must already exist.
      Default = /usr/local/bin ('\${PREFIX}/bin' on Termux for Android)

  -h
      Prints this helpful message and exit."

	echo "${help_header}"
	echo ""
	echo "${help_message}"
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  print_message
#   DESCRIPTION:  Prints a message all fancy like
#    PARAMETERS:  $1 = Message to print
#                 $2 = Severity. info, ok, error, warn
#       RETURNS:  Formatted Message to stdout
#-------------------------------------------------------------------------------
print_message() {
	local message
	local severity
	local red
	local green
	local yellow
	local nc

	message="${1}"
	severity="${2}"
	red='\e[0;31m'
	green='\e[0;32m'
	yellow='\e[1;33m'
	nc='\e[0m'

	case "${severity}" in
	"info") echo -e "${nc}${message}${nc}" ;;
	"ok") echo -e "${green}${message}${nc}" ;;
	"error") echo -e "${red}${message}${nc}" ;;
	"warn") echo -e "${yellow}${message}${nc}" ;;
	esac

}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  make_tempdir
#   DESCRIPTION:  Makes a temp dir using mktemp if available
#    PARAMETERS:  $1 = Directory template
#       RETURNS:  0 = Created temp dir. Also prints temp file path to stdout
#                 1 = Failed to create temp dir
#                 20 = Failed to find mktemp
#-------------------------------------------------------------------------------
make_tempdir() {
	local template
	local tempdir
	local tempdir_rcode

	template="${1}.XXXXXX"

	if command -v mktemp >/dev/null 2>&1; then
		tempdir="$(mktemp -d -t "${template}")"
		tempdir_rcode="${?}"
		if [[ "${tempdir_rcode}" == "0" ]]; then
			echo "${tempdir}"
			return 0
		else
			return 1
		fi
	else
		return 20
	fi
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  determine_os
#   DESCRIPTION:  Attempts to determine host os using uname
#    PARAMETERS:  none
#       RETURNS:  0 = OS Detected. Also prints detected os to stdout
#                 1 = Unknown OS
#                 20 = 'uname' not found in path
#-------------------------------------------------------------------------------
determine_os() {
	local uname_out

	if command -v uname >/dev/null 2>&1; then
		uname_out="$(uname)"
		if [[ "${uname_out}" == "" ]]; then
			return 1
		else
			echo "${uname_out}"
			return 0
		fi
	else
		return 20
	fi
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  determine_arch
#   DESCRIPTION:  Attempt to determine architecture of host
#    PARAMETERS:  none
#       RETURNS:  0 = Arch Detected. Also prints detected arch to stdout
#                 1 = Unknown arch
#                 20 = 'uname' not found in path
#-------------------------------------------------------------------------------
determine_arch() {
	local uname_out

	if command -v uname >/dev/null 2>&1; then
		uname_out="$(uname -m)"
		if [[ "${uname_out}" == "" ]]; then
			return 1
		else
			echo "${uname_out}"
			return 0
		fi
	else
		return 20
	fi
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  download_file
#   DESCRIPTION:  Downloads a file into the specified directory.  Attempts to
#                 use curl, then wget.  If neither is found, fail.
#    PARAMETERS:  $1 = url of file to download
#                 $2 = location to download file into on host system
#       RETURNS:  If curl or wget found, returns the return code of curl or wget
#                 20 = Could not find curl and wget
#-------------------------------------------------------------------------------
download_file() {
	local url
	local dir
	local filename
	local rcode

	url="${1}"
	dir="${2}"
	filename="${3}"

	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "${url}" -o "${dir}/${filename}"
		rcode="${?}"
	elif command -v wget >/dev/null 2>&1; then
		wget --quiet "${url}" -O "${dir}/${filename}"
		rcode="${?}"
	else
		rcode="20"
	fi

	return "${rcode}"
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  checksum_check
#   DESCRIPTION:  Attempt to verify checksum of downloaded file to ensure
#                 integrity.  Tries multiple tools before failing.
#    PARAMETERS:  $1 = path to checksum file
#                 $2 = location of file to check
#                 $3 = working directory
#       RETURNS:  0 = checkusm verified
#                 1 = checksum verification failed
#                 20 = failed to determine tool to use to check checksum
#                 30 = failed to change into or go back from working dir
#-------------------------------------------------------------------------------
checksum_check() {
	local checksum_file
	local file
	local dir
	local rcode
	local shasum_1
	local shasum_2
	local shasum_c

	checksum_file="${1}"
	file="${2}"
	dir="${3}"

	cd "${dir}" || return 30
	if command -v sha256sum >/dev/null 2>&1; then
		## Not all sha256sum versions seem to have --ignore-missing, so filter the checksum file
		## to only include the file we downloaded.
		grep "$(basename "${file}")" "${checksum_file}" >filtered_checksum.txt
		shasum_c="$(sha256sum -c "filtered_checksum.txt")"
		rcode="${?}"
	elif command -v shasum >/dev/null 2>&1; then
		## With shasum on FreeBSD, we don't get to --ignore-missing, so filter the checksum file
		## to only include the file we downloaded.
		grep "$(basename "${file}")" "${checksum_file}" >filtered_checksum.txt
		shasum_c="$(shasum -a 256 -c "filtered_checksum.txt")"
		rcode="${?}"
	elif command -v sha256 >/dev/null 2>&1; then
		## With sha256 on FreeBSD, we don't get to --ignore-missing, so filter the checksum file
		## to only include the file we downloaded.
		## Also sha256 -c option seems to fail, so fall back to an if statement
		grep "$(basename "${file}")" "${checksum_file}" >filtered_checksum.txt
		shasum_1="$(sha256 -q "${file}")"
		shasum_2="$(awk '{print $1}' filtered_checksum.txt)"
		if [[ "${shasum_1}" == "${shasum_2}" ]]; then
			rcode="0"
		else
			rcode="1"
		fi
		shasum_c="Expected: ${shasum_1}, Got: ${shasum_2}"
	else
		return 20
	fi
	cd - >/dev/null 2>&1 || return 30

	if [[ "${rcode}" -gt "0" ]]; then
		echo "${shasum_c}"
	fi
	return "${rcode}"
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  extract_file
#   DESCRIPTION:  Extracts a file into a location.  Attempts to determine which
#                 tool to use by checking file extension.
#    PARAMETERS:  $1 = file to extract
#                 $2 = location to extract file into
#                 $3 = extension
#       RETURNS:  Return code of the tool used to extract the file
#                 20 = Failed to determine which tool to use
#                 30 = Failed to find tool in path
#-------------------------------------------------------------------------------
extract_file() {
	local file
	local dir
	local ext
	local rcode

	file="${1}"
	dir="${2}"
	ext="${3}"

	case "${ext}" in
	"zip")
		if command -v unzip >/dev/null 2>&1; then
			unzip "${file}" -d "${dir}"
			rcode="${?}"
		else
			rcode="30"
		fi
		;;
	"tar.gz")
		if command -v tar >/dev/null 2>&1; then
			tar -xf "${file}" -C "${dir}"
			rcode="${?}"
		else
			rcode="31"
		fi
		;;
	*) rcode="20" ;;
	esac

	return "${rcode}"
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  create_prefix
#   DESCRIPTION:  Creates the install prefix (and any parent directories). If
#                 EUID not 0, then attempt to use sudo.
#    PARAMETERS:  $1 = prefix
#       RETURNS:  Return code of the tool used to make the directory
#                 0 = Created the directory
#                 >0 = Failed to create directory
#                 20 = Could not find mkdir command
#                 21 = Could not find sudo command
#-------------------------------------------------------------------------------
create_prefix() {
	local prefix
	local rcode

	prefix="${1}"

	if command -v mkdir >/dev/null 2>&1; then
		if [[ "${EUID}" == "0" ]]; then
			mkdir -p "${prefix}"
			rcode="${?}"
		else
			if command -v sudo >/dev/null 2>&1; then
				sudo mkdir -p "${prefix}"
				rcode="${?}"
			else
				rcode="21"
			fi
		fi
	else
		rcode="20"
	fi

	return "${rcode}"
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  install_file_linux
#   DESCRIPTION:  Installs a file into a location using 'install'.  If EUID not
#                 0, then attempt to use sudo (unless on android).
#    PARAMETERS:  $1 = file to install
#                 $2 = location to install file into
#       RETURNS:  0 = File Installed
#                 1 = File not installed
#                 20 = Could not find install command
#                 21 = Could not find sudo command
#-------------------------------------------------------------------------------
install_file_linux() {
	local file
	local prefix
	local rcode

	file="${1}"
	prefix="${2}"

	if command -v install >/dev/null 2>&1; then
		if [[ "${EUID}" == "0" ]]; then
			install -C -b -S '_old' -m 755 -t "${prefix}" "${file}"
			rcode="${?}"
		else
			if command -v sudo >/dev/null 2>&1; then
				sudo install -C -b -S '_old' -m 755 "${file}" "${prefix}"
				rcode="${?}"
			elif [[ "${ANDROID_ROOT}" != "" ]]; then
				install -C -b -S '_old' -m 755 -t "${prefix}" "${file}"
				rcode="${?}"
			else
				rcode="21"
			fi
		fi
	else
		rcode="20"
	fi

	return "${rcode}"
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  install_file_freebsd
#   DESCRIPTION:  Installs a file into a location using 'install'.  If EUID not
#                 0, then attempt to use sudo.
#    PARAMETERS:  $1 = file to install
#                 $2 = location to install file into
#       RETURNS:  0 = File Installed
#                 1 = File not installed
#                 20 = Could not find install command
#                 21 = Could not find sudo command
#-------------------------------------------------------------------------------
install_file_freebsd() {
	local file
	local prefix
	local rcode

	file="${1}"
	prefix="${2}"

	if command -v install >/dev/null 2>&1; then
		if [[ "${EUID}" == "0" ]]; then
			install -C -b -B '_old' -m 755 "${file}" "${prefix}"
			rcode="${?}"
		else
			if command -v sudo >/dev/null 2>&1; then
				sudo install -C -b -B '_old' -m 755 "${file}" "${prefix}"
				rcode="${?}"
			else
				rcode="21"
			fi
		fi
	else
		rcode="20"
	fi

	return "${rcode}"
}

#---  FUNCTION  ----------------------------------------------------------------
#          NAME:  main
#   DESCRIPTION:  Put it all together in a logical way
#                 ...at least that is the hope...
#    PARAMETERS:  1 = prefix
#       RETURNS:  0 = All good
#                 1 = Something done broke
#-------------------------------------------------------------------------------
main() {
	local prefix
	local tmpdir
	local tmpdir_rcode
	local dbench_arch
	local dbench_arch_rcode
	local dbench_os
	local dbench_os_rcode
	local dbench_base_url
	local dbench_url
	local dbench_file
	local dbench_checksum_file
	local dbench_bin_name
	local dbench_version
	local dbench_dl_ext
	local download_file_rcode
	local download_checksum_file_rcode
	local checksum_check_rcode
	local extract_file_rcode
	local install_file_rcode
	local create_prefix_rcode
	local bash_autocomplete_file
	local bash_autocomplete_prefix
	local fish_autocomplete_file
	local fish_autocomplete_prefix
	local zsh_autocomplete_file
	local zsh_autocomplete_prefix
	local autocomplete_install_rcode

	dbench_bin_name="dbench"
	dbench_version="0.9.0-alpha"
	dbench_dl_ext="tar.gz"
	dbench_base_url="https://github.com/nikoksr/dbench/releases/download"
	prefix="${1}"
	autocomplete_dir="completions"
	bash_autocomplete_file="${autocomplete_dir}/dbench.bash"
	bash_autocomplete_prefix="/etc/bash_completion.d"
	fish_autocomplete_file="${autocomplete_dir}/dbench.fish"
	fish_autocomplete_prefix="/usr/share/fish/vendor_completions.d"
	zsh_autocomplete_file="${autocomplete_dir}/dbench.zsh"
	zsh_autocomplete_prefix="/etc/zsh"

	print_banner
	print_message "== Install prefix set to ${prefix}" "info"

	tmpdir="$(make_tempdir "${dbench_bin_name}")"
	tmpdir_rcode="${?}"
	if [[ "${tmpdir_rcode}" == "0" ]]; then
		print_message "== Created temp dir at ${tmpdir}" "info"
	elif [[ "${tmpdir_rcode}" == "1" ]]; then
		print_message "== Failed to create temp dir at ${tmpdir}" "error"
	else
		print_message "== 'mktemp' not found in path. Is it installed?" "error"
		exit 1
	fi

	dbench_arch="$(determine_arch)"
	dbench_arch_rcode="${?}"
	if [[ "${dbench_arch_rcode}" == "0" ]]; then
		print_message "== Architecture detected as ${dbench_arch}" "info"
	elif [[ "${dbench_arch_rcode}" == "1" ]]; then
		print_message "== Architecture not detected" "error"
		exit 1
	else
		print_message "== 'uname' not found in path. Is it installed?" "error"
		exit 1
	fi

	dbench_os="$(determine_os)"
	dbench_os_rcode="${?}"
	if [[ "${dbench_os_rcode}" == "0" ]]; then
		print_message "== OS detected as ${dbench_os}" "info"
	elif [[ "${dbench_os_rcode}" == "1" ]]; then
		print_message "== OS not detected" "error"
		exit 1
	else
		print_message "== 'uname' not found in path. Is it installed?" "error"
		exit 1
	fi

	case "${dbench_os}" in
	"Linux") dbench_os="Linux" ;;
	"Darwin") dbench_os="Darwin" ;;
	"CYGWIN"*)
		print_message "== Cygwin is currently unsupported." "error"
		exit 1
		;;
	*)
		print_message "== Unknown OS." "error"
		exit 1
		;;
	esac

	case "${dbench_arch}" in
	"x86_64") dbench_arch="x86_64" ;;
	"amd64") dbench_arch="x86_64" ;;
	"aarch64") dbench_arch="arm64" ;;
	"arm64") dbench_arch="arm64" ;;
	"armv7l") dbench_arch="armv7" ;;
	"i686") dbench_arch="i386" ;;
	*)
		print_message "== Unknown architecture." "error"
		exit 1
		;;
	esac

	dbench_file="${dbench_bin_name}_${dbench_os}_${dbench_arch}.${dbench_dl_ext}"
	dbench_checksum_file="checksums.txt"
	dbench_url="${dbench_base_url}/v${dbench_version}/${dbench_file}"
	dbench_checksum_url="${dbench_base_url}/v${dbench_version}/${dbench_checksum_file}"

	print_message "== Downloading ${dbench_file} from ${dbench_url}" "info"

	download_file "${dbench_url}" "${tmpdir}" "${dbench_file}"
	download_file_rcode="${?}"
	if [[ "${download_file_rcode}" == "0" ]]; then
		print_message "== Downloaded dbench archive into ${tmpdir}" "info"
	elif [[ "${download_file_rcode}" == "1" ]]; then
		print_message "== Failed to download dbench archive" "error"
		exit 1
	elif [[ "${download_file_rcode}" == "20" ]]; then
		print_message "== Failed to locate curl or wget" "error"
		exit 1
	else
		print_message "== Return code of download tool returned an unexpected value of ${download_file_rcode}" "error"
		exit 1
	fi

	download_file "${dbench_checksum_url}" "${tmpdir}" "${dbench_checksum_file}"
	download_checksum_file_rcode="${?}"
	if [[ "${download_checksum_file_rcode}" == "0" ]]; then
		print_message "== Downloaded dbench checksums file into ${tmpdir}" "info"
	elif [[ "${download_checksum_file_rcode}" == "1" ]]; then
		print_message "== Failed to download dbench checksums" "error"
		exit 1
	elif [[ "${download_checksum_file_rcode}" == "20" ]]; then
		print_message "== Failed to locate curl or wget" "error"
		exit 1
	else
		print_message "== Return code of download tool returned an unexpected value of ${download_checksum_file_rcode}" "error"
		exit 1
	fi

	checksum_check "${tmpdir}/${dbench_checksum_file}" "${tmpdir}/${dbench_file}" "${tmpdir}"
	checksum_check_rcode="${?}"
	if [[ "${checksum_check_rcode}" == "0" ]]; then
		print_message "== Checksum of ${tmpdir}/${dbench_file} verified" "ok"
	elif [[ "${checksum_check_rcode}" == "1" ]]; then
		print_message "== Failed to verify checksum of ${tmpdir}/${dbench_file}" "error"
		exit 1
	elif [[ "${checksum_check_rcode}" == "20" ]]; then
		print_message "== Failed to find tool to verify sha256 sums" "error"
		exit 1
	elif [[ "${checksum_check_rcode}" == "30" ]]; then
		print_message "== Failed to change into working directory ${tmpdir}" "error"
		exit 1
	else
		print_message "== Unknown return code returned while checking checksum of ${tmpdir}/${dbench_file}. Returned ${checksum_check_rcode}" "error"
		exit 1
	fi

	extract_file "${tmpdir}/${dbench_file}" "${tmpdir}/" "${dbench_dl_ext}"
	extract_file_rcode="${?}"
	if [[ "${extract_file_rcode}" == "0" ]]; then
		print_message "== Extracted ${dbench_file} to ${tmpdir}/" "info"
	elif [[ "${extract_file_rcode}" == "1" ]]; then
		print_message "== Failed to extract ${dbench_file}" "error"
		exit 1
	elif [[ "${extract_file_rcode}" == "20" ]]; then
		print_message "== Failed to determine which extraction tool to use" "error"
		exit 1
	elif [[ "${extract_file_rcode}" == "30" ]]; then
		print_message "== Failed to find 'unzip' in path" "error"
		exit 1
	elif [[ "${extract_file_rcode}" == "31" ]]; then
		print_message "== Failed to find 'tar' in path" "error"
		exit 1
	else
		print_message "== Unknown error returned from extraction attempt" "error"
		exit 1
	fi

	if [[ ! -d "${prefix}" ]]; then
		create_prefix "${prefix}"
		create_prefix_rcode="${?}"
		if [[ "${create_prefix_rcode}" == "0" ]]; then
			print_message "== Created install prefix at ${prefix}" "info"
		elif [[ "${create_prefix_rcode}" == "20" ]]; then
			print_message "== Failed to find mkdir in path" "error"
			exit 1
		elif [[ "${create_prefix_rcode}" == "21" ]]; then
			print_message "== Failed to find sudo in path" "error"
			exit 1
		else
			print_message "== Failed to create the install prefix: ${prefix}" "error"
			exit 1
		fi
	else
		print_message "== Install prefix already exists. No need to create it." "info"
	fi

	[ ! -d "/etc/bash_completion.d/croc" ] && mkdir -p "/etc/bash_completion.d/croc"
	case "${dbench_os}" in
	"Linux")
		install_file_linux "${tmpdir}/${dbench_bin_name}" "${prefix}/"
		install_file_rcode="${?}"
		;;
	"FreeBSD")
		install_file_freebsd "${tmpdir}/${dbench_bin_name}" "${prefix}/"
		install_file_rcode="${?}"
		;;
	"Darwin")
		install_file_freebsd "${tmpdir}/${dbench_bin_name}" "${prefix}/"
		install_file_rcode="${?}"
		;;
	esac

	if [[ "${install_file_rcode}" == "0" ]]; then
		print_message "== Installed ${dbench_bin_name} to ${prefix}/" "ok"
	elif [[ "${install_file_rcode}" == "1" ]]; then
		print_message "== Failed to install ${dbench_bin_name}" "error"
		exit 1
	elif [[ "${install_file_rcode}" == "20" ]]; then
		print_message "== Failed to locate 'install' command" "error"
		exit 1
	elif [[ "${install_file_rcode}" == "21" ]]; then
		print_message "== Failed to locate 'sudo' command" "error"
		exit 1
	else
		print_message "== Install attempt returned an unexpected value of ${install_file_rcode}" "error"
		exit 1
	fi

	case "$(basename ${SHELL})" in
	"bash")
		install_file_linux "${tmpdir}/${bash_autocomplete_file}" "${bash_autocomplete_prefix}/dbench.bash"
		autocomplete_install_rcode="${?}"
		;;
	"fish")
		install_file_linux "${tmpdir}/${fish_autocomplete_file}" "${fish_autocomplete_prefix}/dbench.fish"
		autocomplete_install_rcode="${?}"
		;;
	"zsh")
		install_file_linux "${tmpdir}/${zsh_autocomplete_file}" "${zsh_autocomplete_prefix}/zsh_autocomplete_dbench"
		autocomplete_install_rcode="${?}"
		print_message "== You will need to add the following to your ~/.zshrc to enable autocompletion" "info"
		print_message "\nPROG=dbench\n_CLI_ZSH_AUTOCOMPLETE_HACK=1\nsource /etc/zsh/zsh_autocomplete_dbench\n" "info"
		;;
	*) autocomplete_install_rcode="1" ;;
	esac

	if [[ "${autocomplete_install_rcode}" == "0" ]]; then
		print_message "== Installed autocompletions for $(basename "${SHELL}")" "ok"
	elif [[ "${autocomplete_install_rcode}" == "1" ]]; then
		print_message "== Failed to install ${bash_autocomplete_file}" "error"
	elif [[ "${autocomplete_install_rcode}" == "20" ]]; then
		print_message "== Failed to locate 'install' command" "error"
	elif [[ "${autocomplete_install_rcode}" == "21" ]]; then
		print_message "== Failed to locate 'sudo' command" "error"
	else
		print_message "== Install attempt returned an unexpected value of ${autocomplete_install_rcode}" "error"
	fi

	print_message "== Installation complete" "ok"

	exit 0
}

#-------------------------------------------------------------------------------
#  ARGUMENT PARSING
#-------------------------------------------------------------------------------
OPTS="hp:"
while getopts "${OPTS}" optchar; do
	case "${optchar}" in
	'h')
		print_help
		exit 0
		;;
	'p')
		INSTALL_PREFIX="${OPTARG}"
		;;
	/?)
		print_message "Unknown option ${OPTARG}" "warn"
		;;
	esac
done

#-------------------------------------------------------------------------------
# CALL MAIN
#-------------------------------------------------------------------------------
main "${INSTALL_PREFIX}"
