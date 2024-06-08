===========
portforward
===========


.. image:: https://img.shields.io/pypi/v/portforward.svg
        :target: https://pypi.python.org/pypi/portforward

.. image:: https://img.shields.io/pypi/status/portforward.svg
        :target: https://pypi.python.org/pypi/portforward

.. image:: https://img.shields.io/pypi/dm/portforward
        :alt: PyPI - Downloads

.. image:: https://readthedocs.org/projects/portforward/badge/?version=latest
        :target: https://portforward.readthedocs.io/en/latest/?version=latest
        :alt: Documentation Status

.. image:: https://github.com/pytogo/portforward/actions/workflows/python-app.yml/badge.svg
        :target: https://github.com/pytogo/portforward/actions
        :alt: Build status



Easy Kubernetes Port-Forward For Python


* Free software: MIT license
* Documentation: https://portforward.readthedocs.io.


Installation
-----------------------------

Wheels are available for:

* Windows (architectures: x64, x86)
* MacOS X (architectures: x86_64, aarch64)
* Linux (architectures: x86_64, x86, aarch64)

with Python versions:

* 3.8
* 3.9
* 3.10
* 3.11
* 3.12

**Requirements for installation from source**

The following things are required when there is no wheel available for the target system.

* `Rust` installed and available in the path (https://www.rust-lang.org/tools/install)
* `Python` (at least v3.7 - below was never tested but might work)

Pip knows how to install ``portforward``.

.. code-block::

    pip install portforward


Quickstart
----------

.. code-block:: Python

    import requests

    import portforward


    def main():
        namespace = "test"
        pod_name = "web"  # You can also use a service name instead
        local_port = 9000  # from port
        pod_port = 80  # to port

        # No path to kube config provided - will use default from $HOME/.kube/config
        with portforward.forward(namespace, pod_name, local_port, pod_port):
            response = requests.get("http://localhost:9000")
            print(f"Done: \n'{response.status_code}'\n'{response.text[:20]}...'")


    if __name__ == "__main__":
        main()


Features
--------

* Native Kubernetes port-forwarding with the ``.kube/config`` from the home dir
  or any other path to config.
* Portforward for pods and services - the lib will first look for a pod with matching name then for
  a service
* Waiting for a pod to become ready
* Multiple forwards per pod or service
* As context manager, sync or async client


Development
-----------

In case you want to develop on this library itself please take a look at the CONTRIBUTING page.

Credits
-------

This project is enabled by PyO3_.

.. _PyO3: https://pyo3.rs
