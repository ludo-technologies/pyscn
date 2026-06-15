class Agg:
    async def sum(self, field, session=None, ignore_cache=False):
        match_stage = self._build_match(session, ignore_cache)
        pipeline = [
            {"$match": match_stage},
            {"$group": {"_id": None, "v": {"$sum": f"${field}"}}},
        ]
        result = await self.aggregate(pipeline, session=session)
        if result and len(result) > 0:
            return result[0]["v"]
        return None

    async def avg(self, field, session=None, ignore_cache=False):
        match_stage = self._build_match(session, ignore_cache)
        pipeline = [
            {"$match": match_stage},
            {"$group": {"_id": None, "v": {"$avg": f"${field}"}}},
        ]
        result = await self.aggregate(pipeline, session=session)
        if result and len(result) > 0:
            return result[0]["v"]
        return None

    async def max(self, field, session=None, ignore_cache=False):
        match_stage = self._build_match(session, ignore_cache)
        pipeline = [
            {"$match": match_stage},
            {"$group": {"_id": None, "v": {"$max": f"${field}"}}},
        ]
        result = await self.aggregate(pipeline, session=session)
        if result and len(result) > 0:
            return result[0]["v"]
        return None

    async def min(self, field, session=None, ignore_cache=False):
        match_stage = self._build_match(session, ignore_cache)
        pipeline = [
            {"$match": match_stage},
            {"$group": {"_id": None, "v": {"$min": f"${field}"}}},
        ]
        result = await self.aggregate(pipeline, session=session)
        if result and len(result) > 0:
            return result[0]["v"]
        return None
