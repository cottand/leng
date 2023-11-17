# Signals (config reload)

Leng will listen for `SIGUSR1` signals in order to reload the config file at run-time
and apply the new config.

Currently, the only field of the config file that supports reloading is
`customdnsrecords`, which specifies custom DNS records served.
Please make an issue if you have a use-case where you would
like some other config fields to also be able to be reloaded.

In the meantime, for all fields, it is safe to simply change the config file while
leng is running, and restart it.

> ### Note on Nomad deployments
>  Nomad is able to send signals rather than restarting a task
> when a template (like the config file) changes.
> 
> It is not recommended to use this approach with leng, because
> there are instances in which Nomad can try to send a signal even if the task
> is not running (see [hashicorp/nomad#5459](https://github.com/hashicorp/nomad/issues/5459),
> which was unresolved at the time of writing).
> 
> Instead, set the template's `change_mode = "restart"` and rely on
> Nomad restarting the task. Due to leng's fast startup time and 
> tiny image size, it should take seconds even when redownloading the image.
> 
> In order to still mitigate this downtime, rolling/canary deployments can
> be used so there is always an instance of leng up to serve traffic.

