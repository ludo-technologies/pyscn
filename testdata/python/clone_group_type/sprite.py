class Sprite:
    def __init__(self, layer, array, array_id):
        self.layer = layer
        self._array = array
        self._array_id = array_id

    def delete(self):
        """Delete this sprite."""
        if self.layer is None:
            return
        self.layer._dirty.discard(self)
        self.layer.objects.discard(self)
        self._array.remove(self._array_id)
        self.layer = None
        self._array = self._array_id = None
