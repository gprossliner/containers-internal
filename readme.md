# Run Command

To test our `run` function, we can execute some commands, like:

* echo hello-world
* sh (interactive)

Test isolation in interactive shell:

- hostname
- ps -aux
- id

# Setup namespaces on Command-Start

Namespaces:
https://man7.org/linux/man-pages/man7/user_namespaces.7.html

Distinct kernel namespace for:
* Unix Timesharing System (hostname)
* Process IDs
* Mounts
* Network
* IPC
* User-ID / Group-ID

Try to change the hostname in sh!

