"""
Inference + SHAP explanation.

score() returns:
  - score  : float in [0, 100] — approval probability × 100
  - shap   : {feature: contribution}  — what the inspector sees
  - flags  : deterministic rule-based annotations for the regulator
"""

import logging
from typing import Any

import numpy as np
import pandas as pd
import shap

from .trainer import FEATURES, ModelTrainer

logger = logging.getLogger(__name__)


class Scorer:
    def __init__(self) -> None:
        self._trainer = ModelTrainer.get_instance()
        self._explainer = shap.TreeExplainer(self._trainer.model)

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def score(
        self, features: dict[str, Any]
    ) -> tuple[float, dict[str, float], list[str]]:
        row = self._build_row(features)

        # Approval probability × 100
        prob = float(self._trainer.model.predict_proba(row[FEATURES])[0][1])
        score = round(prob * 100, 2)

        shap_dict = self._explain(row)
        flags = self._flags(features)

        return score, shap_dict, flags

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    def _build_row(self, f: dict[str, Any]) -> pd.DataFrame:
        return pd.DataFrame(
            [
                {
                    "region_priority": f["region_priority"],
                    "amount_requested": f["amount_requested"],
                    "amount_norm": f["amount_norm"],
                    "amount_ratio": f["amount_ratio"],
                    "month": f["month"],
                    "day_of_year": f["day_of_year"],
                    "farm_size_ha": f["farm_size_ha"],
                    "previous_subsidies_count": f["previous_subsidies_count"],
                    "crop_type_encoded": self._trainer.encode_crop(
                        f.get("crop_type", "other")
                    ),
                }
            ]
        )

    def _explain(self, row: pd.DataFrame) -> dict[str, float]:
        sv = self._explainer.shap_values(row[FEATURES])
        # shap_values() shape varies by SHAP/XGBoost version:
        #   list of 2 arrays (binary, old)  → take index 1
        #   single 2-D array (binary, new)  → take row 0
        if isinstance(sv, list):
            vals = sv[1][0]
        else:
            vals = np.asarray(sv)[0]
        return {feat: round(float(v), 4) for feat, v in zip(FEATURES, vals)}

    @staticmethod
    def _flags(f: dict[str, Any]) -> list[str]:
        """
        Deterministic flags — computed from business rules, NOT from the model.
        This keeps the explanation transparent for regulators: the flag is either
        present or not, with no probabilistic ambiguity.
        """
        flags: list[str] = []
        ratio = f.get("amount_ratio", 0.0)
        if ratio > 1.083:
            excess_pct = (ratio - 1.0) * 100
            flags.append(f"сумма_превышает_норматив_на_{excess_pct:.1f}%")
        if f.get("region_priority", 0) >= 4:
            flags.append("приоритетный_регион")
        if f.get("previous_subsidies_count", 1) == 0:
            flags.append("первичный_заявитель")
        return flags
