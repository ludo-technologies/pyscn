class Parsera:
    def __init__(
        self,
        model: BaseChatModel | None = None,
        extractor: Extractor | None = None,
        initial_script: Callable[[Page], Awaitable[Page]] | None = None,
        stealth: bool = True,
        custom_cookies: list[dict] | None = None,
        typed: bool = False,
    ):
        if model is None and extractor is None:
            self.extractor = APIExtractor()
        elif model and extractor is None:
            self.model = model
            if typed:
                self.extractor = StructuredExtractor(model=self.model)
            else:
                self.extractor = ChunksTabularExtractor(model=self.model)
        elif model is None and extractor:
            self.extractor = extractor
        else:
            raise ValueError(
                "Either model or extractor should be provided, but not both"
            )
        self.initial_script = initial_script
        self.stealth = stealth
        self.loader = PageLoader(custom_cookies=custom_cookies)

    async def _run(
        self,
        url: str,
        elements: dict | None,
        prompt: str,
        proxy_settings: dict | None,
        scrolls_limit: int = 0,
        playwright_script: Callable[[Page], Awaitable[Page]] | None = None,
    ) -> dict:
        if self.loader.context is None:
            await self.loader.create_session(
                proxy_settings=proxy_settings,
                playwright_script=self.initial_script,
                stealth=self.stealth,
            )

        content = await self.loader.fetch_page(
            url=url, scrolls_limit=scrolls_limit, playwright_script=playwright_script
        )

        result = await self.extractor.run(
            content=content, prompt=prompt, attributes=elements
        )
        return result

    def run(
        self,
        url: str,
        elements: dict | None = None,
        prompt: str = "",
        proxy_settings: dict | None = None,
        scrolls_limit: int = 0,
        playwright_script: Callable[[Page], Awaitable[Page]] | None = None,
    ) -> dict:
        return asyncio.run(
            self._run(
                url=url,
                elements=elements,
                prompt=prompt,
                scrolls_limit=scrolls_limit,
                proxy_settings=proxy_settings,
                playwright_script=playwright_script,
            )
        )

    async def arun(
        self,
        url: str,
        elements: dict | None = None,
        prompt: str = "",
        proxy_settings: dict | None = None,
        scrolls_limit: int = 0,
        playwright_script: Callable[[Page], Awaitable[Page]] | None = None,
    ) -> dict:
        return await self._run(
            url=url,
            elements=elements,
            prompt=prompt,
            scrolls_limit=scrolls_limit,
            proxy_settings=proxy_settings,
            playwright_script=playwright_script,
        )
