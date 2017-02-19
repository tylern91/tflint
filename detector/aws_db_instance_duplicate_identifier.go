package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsDBInstanceDuplicateIdentifierDetector struct {
	*Detector
}

func (d *Detector) CreateAwsDBInstanceDuplicateIdentifierDetector() *AwsDBInstanceDuplicateIdentifierDetector {
	return &AwsDBInstanceDuplicateIdentifierDetector{d}
}

func (d *AwsDBInstanceDuplicateIdentifierDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_db_instance") {
		return
	}

	existDBIdentifiers := map[string]bool{}
	resp, err := d.AwsClient.DescribeDBInstances()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}
	for _, dbInstance := range resp.DBInstances {
		existDBIdentifiers[*dbInstance.DBInstanceIdentifier] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_db_instance").Items {
			identifierToken, err := hclLiteralToken(item, "identifier")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			identifier, err := d.evalToString(identifierToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if existDBIdentifiers[identifier] && !d.State.Exists("aws_db_instance", hclObjectKeyText(item)) {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is duplicate identifier. It must be unique.", identifier),
					Line:    identifierToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
