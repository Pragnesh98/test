#include<stdio.h>

int main()
{
	int n1=0;
	char *n2="1234";

	printf("%d\n%s\n",n1,n2);
	n1 = atoi(n2);

	printf("%d\n%s\n",n1,n2);
}
