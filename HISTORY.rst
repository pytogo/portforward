=======
History
=======

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
