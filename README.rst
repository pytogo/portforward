===========
portforward
===========


.. image:: https://img.shields.io/pypi/v/portforward.svg
        :target: https://pypi.python.org/pypi/portforward

.. image:: https://readthedocs.org/projects/portforward/badge/?version=latest
        :target: https://portforward.readthedocs.io/en/latest/?version=latest
        :alt: Documentation Status




Kubernetes Port-Forward Go-Edition For Python


* Free software: MIT license
* Documentation: https://portforward.readthedocs.io.


Features
--------

* Go native Kubernetes port-forwarding with the ``.kube/config`` from the home dir.


How it works
------------

This project uses setuptools-golang_. It will be install through ``pip`` with
the requirements-dev.txt. The following additional lines in the ``setup.py``
activates the compiling:

.. code-block:: Python

    ext_modules=[
        Extension(
            "portforward", ["main.go"],
            py_limited_api=True, define_macros=[('Py_LIMITED_API', None)],
        )
    ]


Credits
-------

This package was created with Cookiecutter_ and the `audreyr/cookiecutter-pypackage`_ project template.

.. _Cookiecutter: https://github.com/audreyr/cookiecutter
.. _`audreyr/cookiecutter-pypackage`: https://github.com/audreyr/cookiecutter-pypackage

This project is enabled by setuptools-golang_.

.. _setuptools-golang: https://github.com/asottile/setuptools-golang
