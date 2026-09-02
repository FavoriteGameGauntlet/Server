package api_generating

//go:generate redocly bundle specification/raw/auth.yaml -o specification/auth.yaml
//go:generate redocly bundle specification/raw/effects.yaml -o specification/effects.yaml
//go:generate redocly bundle specification/raw/games.yaml -o specification/games.yaml
//go:generate redocly bundle specification/raw/exchanges.yaml -o specification/exchanges.yaml
//go:generate redocly bundle specification/raw/items.yaml -o specification/items.yaml
//go:generate redocly bundle specification/raw/points.yaml -o specification/points.yaml
//go:generate redocly bundle specification/raw/perks.yaml -o specification/perks.yaml
//go:generate redocly bundle specification/raw/timers.yaml -o specification/timers.yaml
//go:generate redocly bundle specification/raw/users.yaml -o specification/users.yaml
//go:generate redocly bundle specification/raw/wheel_effects.yaml -o specification/wheel_effects.yaml
//go:generate redocly bundle specification/raw/system_parameters.yaml -o specification/system_parameters.yaml

//go:generate oapi-codegen -config configs/config_auth.yaml specification/auth.yaml
//go:generate oapi-codegen -config configs/config_effects.yaml specification/effects.yaml
//go:generate oapi-codegen -config configs/config_games.yaml specification/games.yaml
//go:generate oapi-codegen -config configs/config_exchanges.yaml specification/exchanges.yaml
//go:generate oapi-codegen -config configs/config_items.yaml specification/items.yaml
//go:generate oapi-codegen -config configs/config_points.yaml specification/points.yaml
//go:generate oapi-codegen -config configs/config_perks.yaml specification/perks.yaml
//go:generate oapi-codegen -config configs/config_timers.yaml specification/timers.yaml
//go:generate oapi-codegen -config configs/config_users.yaml specification/users.yaml
//go:generate oapi-codegen -config configs/config_wheel_effects.yaml specification/wheel_effects.yaml
//go:generate oapi-codegen -config configs/config_system_parameters.yaml specification/system_parameters.yaml
