.. highlight:: shell

============
Installation
============


Stable release
--------------

To install portforward, run this command in your terminal:

.. code-block:: console

    $ pip install portforward

This is the preferred method to install portforward, as it will always install the most recent stable release.

If you don't have `pip`_ installed, this `Python installation guide`_ can guide
you through the process.

.. _pip: https://pip.pypa.io
.. _Python installation guide: http://docs.python-guide.org/en/latest/starting/installation/


From sources
------------

The sources for portforward can be downloaded from the `Github repo`_.

External requirements of this project:

* Golang
* gcc

You can either clone the public repository:

.. code-block:: console

    $ git clone git://github.com/pytogo/portforward

Or download the `tarball`_:

.. code-block:: console

    $ curl -OJL https://github.com/pytogo/portforward/tarball/master

Once you have a copy of the source and the external dependencies ready, you can install it with:

.. code-block:: console

    $ python setup.py install


.. _Github repo: https://github.com/pytogo/portforward
.. _tarball: https://github.com/pytogo/portforward/tarball/master
