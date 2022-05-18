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

# CLONE_NEWUSER (>= Linux 3.8) 

>      Starting in Linux 3.8, unprivileged processes can create user
       namespaces, and the other types of namespaces can be created with
       just the CAP_SYS_ADMIN capability in the caller's user namespace.

       When a nonuser namespace is created, it is owned by the user
       namespace in which the creating process was a member at the time
       of the creation of the namespace.  Privileged operations on
       resources governed by the nonuser namespace require that the
       process has the necessary capabilities in the user namespace that
       owns the nonuser namespace.

       If CLONE_NEWUSER is specified along with other CLONE_NEW* flags
       in a single clone(2) or unshare(2) call, the user namespace is
       guaranteed to be created first, giving the child (clone(2)) or
       caller (unshare(2)) privileges over the remaining namespaces
       created by the call.  Thus, it is possible for an unprivileged
       caller to specify this combination of flags.

* Let's checkout the code with CLONE_NEWUSER!
* And try to change the hostname! -> peng...


# UID (User-ID) mapping

Allows to 

>    User and group ID mappings: uid_map and gid_map
       When a user namespace is created, it starts out without a mapping
       of user IDs (group IDs) to the parent user namespace.  The
       /proc/[pid]/uid_map and /proc/[pid]/gid_map files (available
       since Linux 3.5) expose the mappings for user and group IDs
       inside the user namespace for the process pid.  These files can
       be read to view the mappings in a user namespace and written to
       (once) to define the mappings.

This allows us to map id 0 (root) in container to be XXX (unprivileged) outside container

* Check the setup of the uid_map in code
* Check /proc/XXX/uid_map (`echo $$` to get current PID)

# Process-ID Namespace

By specifing the `CLONE_NEWPID` flag, you can create a new PID namespace.

* Check the code for the added flag
* `echo $$` in shell
* `ps -x` in shell -> ?

