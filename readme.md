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

# Give it an own root

Download an Alpine "Mini Root Filesystem" for our container

https://www.alpinelinux.org/downloads/
https://dl-cdn.alpinelinux.org/alpine/v3.15/releases/x86_64/alpine-minirootfs-3.15.4-x86_64.tar.gz

```sh
pushd /tmp
mkdir alpineroot
cd alpineroot
wget https://dl-cdn.alpinelinux.org/alpine/v3.15/releases/x86_64/alpine-minirootfs-3.15.4-x86_64.tar.gz
tar -xvf alpine-minirootfs-3.15.4-x86_64.tar.gz
```

Use `chroot` (https://linux.die.net/man/1/chroot) to specify a new root `/` for the container.

* The code has to be modified a bit, because we can't pass the new root when starting a process (like NS).
* We need to invoke the `chroot` before invoking our command
* But we don't wanna mess around with the state of our main process
* So we create a process tree main -> child -> run

Test running the following:
* `ls /`
* `ps -a`

# Mount proc

We have a distinct mount ns (`CLONE_NEWNS`), but have not mounted anything
If you check `/proc/??/mounts` you'll the system mounts
If you check the Alpine Root Image, you see /proc in an empty dir (prepared for mounting)

Check the Code for the `syscall.Mount` of the `proc` filesystem!
And test `ps -a` again!

We chould do this for /tmp, /dev, ... etc.

# Use overlayfs

All changes to the filesystem are written directly to our Alpine root.
So we can't isolate container instances.

For such an isolation, we use a differenting filesystem, like overlayfs.
https://www.kernel.org/doc/html/latest/filesystems/overlayfs.html

> This document describes a prototype for a new approach to providing overlay-filesystem functionality in Linux (sometimes referred to as union-filesystems). An overlay-filesystem tries to present a filesystem which is the result over overlaying one filesystem on top of the other.

The "lowerdir" is the readonly baseline. For multiple layers, you can have multiple lowerdirs
The "upperdir" is the writable layer above. All changes to the resulting filesystem are written to the "upperdir".

Our setup:

lowerdir=Alpine root
upperdir=tempoary location, this will contain all changes we make to the fs
workdir=tempoary location, this is needed by overlay for some operations (like atomic move)

* Check the code for our overlay fs
* Try to create new files, and check upperdir
* Try to delete existing file, and check upperdir
* Restart to start from scratch
* Anaylse docker images, containers, and overlayfs

