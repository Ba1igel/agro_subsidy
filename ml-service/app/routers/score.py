import logging
from functools import lru_cache

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from app.model.scorer import Scorer

logger = logging.getLogger(__name__)
router = APIRouter()


# ------------------------------------------------------------------
# Schemas
# ------------------------------------------------------------------


class ScoreRequest(BaseModel):
    task_id: str
    region_priority: int = Field(..., ge=1, le=5)
    amount_requested: float = Field(..., gt=0)
    amount_norm: float = Field(..., gt=0)
    amount_ratio: float = Field(..., ge=0)
    month: int = Field(..., ge=1, le=12)
    day_of_year: int = Field(..., ge=1, le=366)
    crop_type: str
    farm_size_ha: float = Field(..., gt=0)
    previous_subsidies_count: int = Field(..., ge=0)


class ScoreResponse(BaseModel):
    task_id: str
    score: float
    shap_values: dict[str, float]
    flags: list[str]


# ------------------------------------------------------------------
# Dependency — singleton scorer, built once after model is trained
# ------------------------------------------------------------------


@lru_cache(maxsize=1)
def _get_scorer() -> Scorer:
    return Scorer()


# ------------------------------------------------------------------
# Endpoint
# ------------------------------------------------------------------


@router.post("/score", response_model=ScoreResponse)
def score_task(req: ScoreRequest) -> ScoreResponse:
    try:
        scorer = _get_scorer()
        score, shap_vals, flags = scorer.score(req.model_dump())
        return ScoreResponse(
            task_id=req.task_id,
            score=score,
            shap_values=shap_vals,
            flags=flags,
        )
    except Exception as exc:
        logger.exception("Scoring failed for task %s", req.task_id)
        raise HTTPException(status_code=500, detail=str(exc)) from exc
