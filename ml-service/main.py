import logging

from fastapi import FastAPI

from app.model.trainer import ModelTrainer
from app.routers import score

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s — %(message)s",
)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Agro Subsidy ML Service",
    description="XGBoost scoring + SHAP explanations for agricultural subsidy applications.",
    version="1.0.0",
)

app.include_router(score.router)


@app.on_event("startup")
async def startup() -> None:
    logger.info("Warming up model…")
    ModelTrainer.get_instance()
    logger.info("Model ready — service accepting requests.")


@app.get("/health", tags=["ops"])
def health() -> dict:
    return {"status": "ok"}
