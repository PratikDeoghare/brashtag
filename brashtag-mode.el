(define-minor-mode brashtag-mode "Minor mode for brashtag documents."
  :lighter " BT"
  (message "Entering brashtag mode..")
  (text-mode)
  (unhighlight-regexp "\\(#.*?\\){")
  (highlight-regexp "\\(#.*?\\){" 'font-lock-constant-face 1)
)


