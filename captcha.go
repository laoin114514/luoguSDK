package luogusdk

import "laoin114514/luoguSDK/internal/ocr"

// DDDDOCRSolver 返回一个使用 ddddocr 的 CaptchaSolver
func DDDDOCRSolver() CaptchaSolver {
	engine := ocr.NewDDddocr()
	return func(image []byte) (string, error) {
		return engine.Recognize(image)
	}
}

// DDDDOCRSolverWithPython 指定 Python 解释器路径
func DDDDOCRSolverWithPython(python string) CaptchaSolver {
	engine := ocr.NewDDddocrWithPython(python)
	return func(image []byte) (string, error) {
		return engine.Recognize(image)
	}
}
