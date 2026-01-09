/*
 * Copyright © 2025 Musing Studio LLC.
 *
 * This file is part of WriteFreely.
 *
 * WriteFreely is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, included
 * in the LICENSE file in this source code package.
 */

package writefreely

import (
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/microcosm-cc/bluemonday"
)

var htmlSanitizer = bluemonday.UGCPolicy()

func HTMLToMarkdown(html string) string {
	sanitized := htmlSanitizer.Sanitize(html)

	md, err := htmltomarkdown.ConvertString(sanitized)
	if err != nil {
		return sanitized
	}

	return md
}
