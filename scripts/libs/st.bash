#!/bin/bash

bold=$(tput smso)
offbold=$(tput rmso)
resetColor="$(tput sgr0)"
red="$(
    tput bold
    tput setaf 1
)"
green="$(
    tput bold
    tput setaf 2
)"
yellow="$(
    tput bold
    tput setaf 3
)"
blue="$(
    tput bold
    tput setaf 4
)"
blueCyan="$(
    tput bold
    tput setaf 6
)"

DOING_MSG=

function st.h1() {
    echo "st.h1> ${bold}$1${offbold}"
}

function st.h2() {
    echo "st.h2> \t${bold}$1${offbold}"
}

function st.h3() {
    echo "st.h3> \t\t${bold}$1${offbold}"
}

function st.doing() {
    DOING_MSG=$1

    echo "st.doing> ${blue} Doing « ${DOING_MSG} »…$resetColor"
}

function st.done() {
    local DONE="${1:-DONE}"

    echo
    echo "st.done> ${DOING_MSG} : ${green}$DONE${resetColor}"
    echo
}

function st.nothingTodo() {
    _done 'st.nothingtd> Nothing to do…'
}

function st.skipped() {
    echo
    echo "st.skiped> ${DOING_MSG} : ${blueCyan}SKIPPED${resetColor}"
    echo
}

function st.warn() {
    echo
    echo "st.warn> ${DOING_MSG} : ${bold}${yellow}$1${resetColor}${offbold}"
    echo
}

function st.fail() {
    echo "st.fail> $bold${red}PROCESS ABORTED$resetColor$offbold"
    echo

    exit 1
}

function st.do() {
    # https://unix.stackexchange.com/questions/148109/shifting-command-output-to-the-right
    local -a cmd=("$@")
    echo "st.do> ${blueCyan}${cmd[@]}${resetColor}"

    "${cmd[@]}" || _fail | nl -bn
}
