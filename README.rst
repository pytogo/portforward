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



Kubernetes Port-Forward Go-Edition For Python


* Free software: MIT license
* Documentation: https://portforward.readthedocs.io.


Installation
-----------------------------

Wheels are available for:

* Windows
* MacOS X
* Linux

with Python versions:

* 3.6
* 3.7
* 3.8
* 3.9
* 3.10

and architectures:

* x84_64
* arm64 (known as M1/Apple Chip - MacOS only)

**Requirements for installation from source**

The following things are required when there is no wheel available for the target system.

* `Go` installed and available in the path (at least v1.16 / https://go.dev)
* `Python` (at least v3.6 - below was never tested but might work)
* `gcc` (for Windows available via MinGW)

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
        pod_name = "web"
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

* Go native Kubernetes port-forwarding with the ``.kube/config`` from the home dir
  or any other path to config.


Credits
-------

This package was created with Cookiecutter_ and the `audreyr/cookiecutter-pypackage`_ project template.

.. _Cookiecutter: https://github.com/audreyr/cookiecutter
.. _`audreyr/cookiecutter-pypackage`: https://github.com/audreyr/cookiecutter-pypackage

This project is enabled by setuptools-golang_.

.. _setuptools-golang: https://github.com/asottile/setuptools-golang
