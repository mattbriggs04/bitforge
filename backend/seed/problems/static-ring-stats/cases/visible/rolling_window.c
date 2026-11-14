stats_reset();
for (uint16_t i = 1; i <= 9; i++) {
    stats_record(i);
}
case_passed = stats_count() == 8 &&
              stats_min() == 2 &&
              stats_max() == 9 &&
              stats_avg() == 5;
