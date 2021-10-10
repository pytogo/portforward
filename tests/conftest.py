import sys
from unittest.mock import MagicMock


def _mock_forward(*_args, **_kwargs):
    raise RuntimeError("Go portforward exploded")


# The tests should not test the internal _portforward module
_portforward = MagicMock()
_portforward.forward = _mock_forward
sys.modules["_portforward"] = _portforward
