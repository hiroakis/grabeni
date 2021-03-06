package aws

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Return client for test
func newClient(svc *EC2API) *ENIClient {
	f, _ := os.Open(os.DevNull)
	l := log.New(f, "", 0)
	return &ENIClient{
		svc:       svc,
		logger:    l,
		logWriter: new(bytes.Buffer),
	}
}

func TestDescribeENIByID(t *testing.T) {
	{
		mockEC2 := new(EC2API)
		c := newClient(mockEC2)

		mockEC2.On("DescribeNetworkInterfaces", &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String("eni-00000001")},
		}).Return(&ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []*ec2.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-00000001"),
					Attachment: &ec2.NetworkInterfaceAttachment{
						InstanceId: aws.String("i-00000001"),
					},
				},
			},
		}, nil)

		mockEC2.On("DescribeInstances", &ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String("i-00000001")},
		}).Return(&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				{
					Instances: []*ec2.Instance{{
						InstanceId: aws.String("i-00000001"),
					}},
				},
			},
		}, nil)

		eni, err := c.DescribeENIByID("eni-00000001")

		assert.NoError(t, err)
		assert.NotNil(t, eni.AttachedInstance())
		assert.Equal(t, "i-00000001", eni.AttachedInstanceID())
	}

	{
		mockEC2 := new(EC2API)
		c := newClient(mockEC2)

		mockEC2.On("DescribeNetworkInterfaces", &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String("eni-00000001")},
		}).Return(&ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []*ec2.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-00000001"),
					// Not attached
				},
			},
		}, nil)

		mockEC2.On("DescribeInstances", &ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String("i-00000001")},
		}).Return(&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				{
					Instances: []*ec2.Instance{{
						InstanceId: aws.String("i-00000001"),
					}},
				},
			},
		}, nil)

		eni, err := c.DescribeENIByID("eni-00000001")

		assert.NoError(t, err)
		assert.Nil(t, eni.AttachedInstance())
	}
}

func TestDescribeENIs(t *testing.T) {
	mockEC2 := new(EC2API)
	c := newClient(mockEC2)

	mockEC2.On("DescribeNetworkInterfaces", mock.Anything).Return(
		&ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []*ec2.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-00000001"),
					Attachment: &ec2.NetworkInterfaceAttachment{
						InstanceId: aws.String("i-00000001"),
					},
				},
				{
					NetworkInterfaceId: aws.String("eni-00000002"),
					Attachment:         nil,
				},
			},
		}, nil)

	mockEC2.On("DescribeInstances", &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String("i-00000001")},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{{
					InstanceId: aws.String("i-00000001"),
				}},
			},
		},
	}, nil)

	enis, err := c.DescribeENIs()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(enis))
	if assert.NotNil(t, enis[0].AttachedInstance()) {
		assert.Equal(t, "i-00000001", enis[0].AttachedInstanceID())
	}
	assert.Nil(t, enis[1].AttachedInstance())
}

func TestDescribeInstanceByID(t *testing.T) {
	mockEC2 := new(EC2API)
	c := newClient(mockEC2)

	mockEC2.On("DescribeInstances", &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String("i-00000001")},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{{
					InstanceId: aws.String("i-00000001"),
				}},
			},
		},
	}, nil)

	i, err := c.DescribeInstanceByID("i-00000001")

	assert.NoError(t, err)
	assert.Equal(t, "i-00000001", i.InstanceID())
}

func TestDescribeInstancesByID(t *testing.T) {
	mockEC2 := new(EC2API)
	c := newClient(mockEC2)

	mockEC2.On("DescribeInstances", &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String("i-00000001")},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{{
					InstanceId: aws.String("i-00000001"),
				}},
			},
		},
	}, nil)

	instances, err := c.DescribeInstancesByIDs([]string{"i-00000001"})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(instances))
	assert.Equal(t, "i-00000001", instances[0].InstanceID())

	mockEC2 = new(EC2API)
	c = newClient(mockEC2)

	mockEC2.On("DescribeInstances", &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String("i-00000000")},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{Instances: nil},
		},
	}, nil)

	instances, err = c.DescribeInstancesByIDs([]string{"i-00000000"})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(instances))
}
