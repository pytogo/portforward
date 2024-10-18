=======
History
=======

0.7.0 (2024-10-18)
------------------
* Allow binding to a local random free port

0.6.2 (2024-06-19)
------------------
* Allow defining binding ip

0.6.1 (2024-01-25)
------------------
* Fixed wrong logger setup

0.6.0 (2023-06-13)
------------------
* Rewrite native part in Rust with support of Py03 and maturin
* Add async client

0.5.0 (2023-05-19)
------------------
* Move pytogo Go code into portforward
* Fix stopping portforward for services
* Allow portforwarding without contextmanager
* Allow multiple portforwards to same pod or service

0.4.5 (2023-03-06)
------------------
* Fix panic when logging an error
* Change default log level to INFO

0.4.4 (2023-02-28)
------------------
* Fix endless waiting

0.4.3 (2023-02-27)
------------------
* Throw error instead of panic when port is in usage

0.4.2 (2023-02-06)
------------------
* Use in-cluster-config when no kube config file is available

0.4.1 (2023-02-01)
------------------
* Bump pytogo/portforward version

0.4.0 (2023-01-31)
------------------
* Respect environment variable KUBECONFIG
* Wait if a pod is not ready yet
* Be able to use service as targets

0.3.1 (2022-12-26)
------------------
* Allow selecting kubernetes target context

0.3.0 (2022-10-08)
------------------
* Introduction of logging level as replacement for verbose mode


0.2.8 (2022-08-22)
------------------
* Added verbose mode


0.2.7 (2021-10-05)
------------------
* Added missing import
* Added type hint


0.2.6 (2021-10-05)
------------------
* Fixed type hint


0.2.5 (2021-09-09)
------------------
* Moved the actual portforward to own module


0.2.4 (2021-08-23)
------------------
* Added adal import for Azure AD
* Fixed host IPs with paths
* Made timeout flexible


0.2.3 (2021-08-23)
------------------
* Fixed case when hostIP contains a path
* Added common and cloud provider auth plugins


0.2.2 (2021-08-23)
------------------
* Fixed missing module ``portforward``


0.2.1 (2021-08-19)
------------------
* Decrease binary size if pre-compile wheels
  (Improvement of setuptools-golang)


0.2.0 (2021-08-14)
------------------

* First Release on PyPI.
* Made path to kube config variable.
* Port-forwarding became non-blocking.
* Fixed verification bug when port was None.
* Added throwing own error.


0.1.0 (2021-08-09)
------------------

* First release on Test PyPI.
* Blocking port-forward with fixed path for kube config.
