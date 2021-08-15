===========
portforward
===========


.. image:: https://img.shields.io/pypi/v/portforward.svg
        :target: https://pypi.python.org/pypi/portforward

.. image:: https://img.shields.io/pypi/status/portforward.svg
        :target: https://pypi.python.org/pypi/portforward

.. image:: https://readthedocs.org/projects/portforward/badge/?version=latest
        :target: https://portforward.readthedocs.io/en/latest/?version=latest
        :alt: Documentation Status




Kubernetes Port-Forward Go-Edition For Python


* Free software: MIT license
* Documentation: https://portforward.readthedocs.io.


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
