# Installer
A small golang app for executing installer scripts.

The purpose of this thing is to keep all of the small apps I have installed outside
of my distros repo up-to-date. I do not like flatpak, so I made my own ;)

This "installer" simply replicates the manual process of checking the local
version vs the remote version and if needed download, verify and install a new version.

The actual scripts are stored in jsol/installer-scripts and the path to that repo
(or your own version) must be added to the environment variable INSTALLER_BASEDIR.

Structure of a set of install scripts is
<name> - A directory with the name of the app
  ├─ local-version.sh - Must output the local version in the same style as the remote version
  ├─ remote-version.sh - Output the newest version on the "remote" (usually github release)
  ├─ dependencies.sh   - Optional file for checking that all needed dependencies are present
  ├─ download.sh       - Download a specified file
  ├─ install.sh        - Install the downloaded file (ran as root with sudo)


  Run ./installer install <name> for a new install of <name>
  Run ./installer update  to check all installed apps for updates

## Github rate limiting
If you are on a shared IP on say a corporate net then GitHub probably ratelimits you.
Install the github CLI and do a gh login, then that token will be used to authenticate
and thus bypass the rate-limit.

## TODOs
Maybe store the previous version as well, enabling a "rollback" if the new one fails
There is probably a ton of improvements to be made in ergonomics and safety. Use at your own risk.



