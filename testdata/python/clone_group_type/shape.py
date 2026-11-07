class Shape:
    def __init__(self, layer, vao, lst):
        self.layer = layer
        self.vao = vao
        self.lst = lst

    def delete(self):
        """Delete this primitive."""
        if self.layer is None:
            return
        self.layer._dirty.discard(self)
        self.layer.objects.discard(self)
        self.vao.free(self.lst)
        self.lst = None
        self.layer = None
