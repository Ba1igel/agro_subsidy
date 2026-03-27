"""
Feature engineering + XGBoost training on synthetic data.

Singleton pattern: ModelTrainer.get_instance() trains once at startup
and is reused for all inference calls.
"""

import logging

import numpy as np
import pandas as pd
from sklearn.preprocessing import LabelEncoder
from xgboost import XGBClassifier

logger = logging.getLogger(__name__)

# Ordered feature list — must match scorer.py
FEATURES = [
    "region_priority",
    "amount_requested",
    "amount_norm",
    "amount_ratio",
    "month",
    "day_of_year",
    "farm_size_ha",
    "previous_subsidies_count",
    "crop_type_encoded",
]

CROP_TYPES = ["wheat", "corn", "sunflower", "barley", "soy", "other"]


class ModelTrainer:
    _instance: "ModelTrainer | None" = None

    def __init__(self) -> None:
        self.label_encoder = LabelEncoder()
        self.label_encoder.fit(CROP_TYPES)
        self.model: XGBClassifier = self._train()

    @classmethod
    def get_instance(cls) -> "ModelTrainer":
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    # ------------------------------------------------------------------
    # Training
    # ------------------------------------------------------------------

    def _train(self) -> XGBClassifier:
        logger.info("Generating synthetic training data (n=5000)…")
        rng = np.random.default_rng(42)
        n = 5_000

        df = pd.DataFrame(
            {
                "region_priority": rng.integers(1, 6, n),
                "amount_requested": rng.uniform(10_000, 500_000, n),
                "amount_norm": rng.uniform(50_000, 300_000, n),
                "month": rng.integers(1, 13, n),
                "day_of_year": rng.integers(1, 366, n),
                "farm_size_ha": rng.uniform(10, 5_000, n),
                "previous_subsidies_count": rng.integers(0, 11, n),
                "crop_type_encoded": rng.integers(0, len(CROP_TYPES), n),
            }
        )
        df["amount_ratio"] = df["amount_requested"] / df["amount_norm"]

        # Interpretable synthetic label: high priority + low ratio + big farm → approve
        raw_score = (
            df["region_priority"] * 15
            + np.clip((1.0 - df["amount_ratio"]) * 30, -30, 30)
            + np.log1p(df["farm_size_ha"]) * 2
            + df["previous_subsidies_count"] * 3
            + np.where(df["month"].between(3, 9), 5, 0)
        )
        y = (raw_score > np.median(raw_score)).astype(int)

        model = XGBClassifier(
            n_estimators=100,
            max_depth=4,
            learning_rate=0.1,
            eval_metric="logloss",
            random_state=42,
        )
        model.fit(df[FEATURES], y)
        logger.info("XGBoost model trained successfully.")
        return model

    # ------------------------------------------------------------------
    # Helpers
    # ------------------------------------------------------------------

    def encode_crop(self, crop_type: str) -> int:
        crop = crop_type.lower() if crop_type.lower() in CROP_TYPES else "other"
        return int(self.label_encoder.transform([crop])[0])
